package graph

import (
	"bytes"
	"context"
	"errors"
	"flag"
	"io"
	"io/fs"
	"path/filepath"
	"log/slog"
	"os"
	"strings"
	"fmt"

	"gopkg.in/yaml.v3"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"

	"github.boschdevcloud.com/fsil/fsil.go/command"

	"github.com/boschglobal/dse.sdp/graph/internal/pkg/graph"
)

type GraphReportCommand struct {
	command.Command
	optTag      string
	optDb       string
	reportFile  string
}

type Query struct {
	Name     string `yaml:"name"`
	Evaluate bool   `yaml:"evaluate,omitempty"`
	Query    string `yaml:"query"`
}

type Report struct {
	Name  	string     `yaml:"name"`
	Tags    []string   `yaml:"tags"`
	Queries []Query    `yaml:"queries"`
	Hint  	string     `yaml:"hint"`
}

func NewGraphReportCommand(name string) *GraphReportCommand {
	c := &GraphReportCommand{
		Command: command.Command{
			Name:    name,
			FlagSet: flag.NewFlagSet(name, flag.ExitOnError),
		},
	}
	c.FlagSet().StringVar(&c.optTag, "tag", "", "run all reports with specified tag")
	c.FlagSet().StringVar(&c.optDb, "db", "bolt://localhost:7687", "database connection string")
	return c
}

// CommandRunner interface functions.
func (c GraphReportCommand) Name() string {
	return c.Command.Name
}

func (c GraphReportCommand) FlagSet() *flag.FlagSet {
	return c.Command.FlagSet
}

func (c *GraphReportCommand) Parse(args []string) error {
	err := c.FlagSet().Parse(args)
    if err != nil {
        return err
    }
    if c.FlagSet().NArg() != 1 {
        return fmt.Errorf("report file not specified")
    }
    c.reportFile = c.FlagSet().Arg(0)
    return nil
}

func (c *GraphReportCommand) Run() error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %w", err)
	}

	var paths []string
	onlyCheckCurrentDir := false

	// If reportFile starts with "./", only check current directory.
	if strings.HasPrefix(c.reportFile, "./") {
		paths = []string{filepath.Join(".", c.reportFile)}
		onlyCheckCurrentDir = true
	} else {
		paths = []string{
			filepath.Join(homeDir, ".local", "share", "dse-graph", "reports", c.reportFile),
			filepath.Join(".", c.reportFile),
		}
	}

	var reportPath string
	for _, path := range paths {
		info, err := os.Stat(path)
		if err == nil && !info.IsDir() {
			reportPath = path
			break
		}
	}

	if reportPath == "" {
		if onlyCheckCurrentDir {
			return fmt.Errorf("report file %q not found in current directory", c.reportFile)
		}
		return fmt.Errorf("report file %q not found in local share or current directory", c.reportFile)
	}

    slog.Info("Connecting to graph", "db", c.optDb)
    ctx := context.Background()
    driver, err := graph.Driver(c.optDb)
    if err != nil {
        slog.Error("Graph driver error", "error", err)
        return err
    }
    ctx = context.WithValue(ctx, "driver", driver)
    defer graph.Close(ctx)

    session, err := graph.Session(ctx)
    if err != nil {
        slog.Error("Graph session error", "error", err)
        return err
    }
    defer session.Close(ctx)

    args := c.FlagSet().Args()
    if len(args) == 0 {
        slog.Info("Usage: graph report <yaml-file> OR graph report -tag <tag-name> <folder|file>")
        return nil
    }

    // Process the Generated Report.
    c.processReports(ctx, session, reportPath)

	return nil
}

func (c *GraphReportCommand) processReports(ctx context.Context, session neo4j.SessionWithContext, fileOrFolder string) error {
	fileInfo, err := os.Stat(fileOrFolder)
	if err != nil {
		slog.Error("Error accessing file/folder", "error", err)
		return err
	}

	// Scan through subdirectories and process all YAML Report files.
	if fileInfo.IsDir() {
		err := filepath.WalkDir(fileOrFolder, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				slog.Error("Error accessing path", "path", path, "error", err)
				return err
			}
			if !d.IsDir() && (strings.HasSuffix(path, ".yaml") || strings.HasSuffix(path, ".yml")) {
				return c.processReports(ctx, session, path)
			}
			return nil
		})
		return err
	}

	// Process a single YAML file.
	fileData, err := os.ReadFile(fileOrFolder)
	if err != nil {
		slog.Error("Error reading YAML file", "file", fileOrFolder, "error", err)
		return err
	}

	decoder := yaml.NewDecoder(bytes.NewReader(fileData))

	var totalReports, passedReports, failedReports int
	var report Report
	var failedList []string

	for {
		// Decode next YAML document
		if err := decoder.Decode(&report); err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			slog.Error("Error parsing YAML", "error", err)
			return err
		}

		// Skip YAML files without a query field.
		if len(report.Queries) == 0 {
			slog.Warn("Skipping non-Report YAML file.", "file", fileOrFolder)
			continue
		}

		totalReports++

		// Check if tag filtering is enabled.
		if c.optTag != "" {
			tagSet := make(map[string]struct{})
			for _, tag := range report.Tags {
				tagSet[tag] = struct{}{}
			}
			reqTags := strings.Split(c.optTag, ",")
			tagMatch := false
			for _, reqTag := range reqTags {
				if _, exists := tagSet[reqTag]; exists {
					tagMatch = true
					break
				}
			}
			if !tagMatch {
				continue // Skip if no matching tag.
			}
		}

		// Run query for each YAML report.
		if err := c.runReport(ctx, session, fileOrFolder, report); err != nil {
			failedReports ++
			failedList = append(failedList, report.Name)
		} else {
			passedReports ++
		}
	}

	// Print summary.
	summary := fmt.Sprintf("\n=================== Summary ===================\nRan %d Reports | Passed: %d | Failed: %d\n",
    passedReports+failedReports, passedReports, failedReports)
	if len(failedList) > 0 {
		summary += "Failed Reports: " + strings.Join(failedList, ", ")
	}
	summary += "\n===============================================\n"
	fmt.Println(summary)

	if failedReports > 0 {
		os.Exit(1)
	}

	return nil
}

func (c *GraphReportCommand) runReport(ctx context.Context, session neo4j.SessionWithContext, fileOrFolder string, report Report) error {
	fmt.Println()
	slog.Info(fmt.Sprintf("Report name: %s", report.Name))
	slog.Info(fmt.Sprintf("Path to Report: %s", fileOrFolder))

	// Check if there are queries.
	if len(report.Queries) == 0 {
		slog.Info("No queries found in report YAML", "file", fileOrFolder)
		return fmt.Errorf("no queries found in YAML file")
	}

	// Execute each query and print individual tables.
	for _, q := range report.Queries {
		slog.Info(fmt.Sprintf("Query Name: %s: \n%s\n", q.Name, q.Query))

		// Execute query.
		result, err := session.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
			queryResult, err := tx.Run(ctx, q.Query, nil)
			if err != nil {
				return nil, err
			}
			return queryResult.Collect(ctx)
		})
		if err != nil {
			slog.Error("Failed to execute query", "error", err)
			continue
		}

		// Type assertion to ensure result is of expected type.
		records, ok := result.([]*neo4j.Record)
		if !ok {
			slog.Error("Unexpected result type from query execution")
			continue
		}

		// Check if result is empty.
		if len(records) == 0 {
			slog.Info("Query returned no results")
			return fmt.Errorf("Report Failed")
		}

		// Print the table with results.
		printTable(records)
		fmt.Println()

		if q.Evaluate {
			// Determine pass/fail based on result value.
			for _, record := range records {
				resultValue, _ := record.Get("result")
				if resultValue != "PASS" {
					// If evaluation fails, provide hint.
					slog.Info(fmt.Sprintf("Hint !! %s", report.Hint))
					fmt.Println()
					slog.Info("Report Failed")
					return fmt.Errorf("Report Failed")
					break
				}
			}

			slog.Info("Report Passed\n\n")

		}
	}

	fmt.Println(strings.Repeat("=", 100))
	fmt.Println()
	fmt.Println(strings.Repeat("=", 100))

	return nil
}

// Print results in a simple table.
func printTable(records []*neo4j.Record) {
	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)

	// Table header.
	headers := records[0].Keys
	rowHeaders := make([]interface{}, len(headers))
	for i, col := range headers {
		rowHeaders[i] = col
	}
	t.AppendHeader(rowHeaders)

	// Table rows.
	for _, record := range records {
		row := make([]interface{}, len(headers))
		for i, col := range headers {
			value, _ := record.Get(col)
			row[i] = value
		}
		t.AppendRow(row)
	}
	t.Render()
}

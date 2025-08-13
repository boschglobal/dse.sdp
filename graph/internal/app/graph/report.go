package graph

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
	"gopkg.in/yaml.v3"

	"github.com/boschglobal/dse.clib/extra/go/command"
	"github.com/boschglobal/dse.clib/extra/go/command/log"

	"github.com/boschglobal/dse.sdp/graph/internal/pkg/file/kind"
	"github.com/boschglobal/dse.sdp/graph/internal/pkg/graph"
)

type GraphReportCommand struct {
	command.Command
	logLevel    int
	optTags     []string
	optDb       string
	optNames    []string
	optReport   string
	optList     bool
	optListTags bool
	optListAll  bool
	reportFile  string
}

type Query struct {
	Name       string `yaml:"name"`
	Evaluate   bool   `yaml:"evaluate,omitempty"`
	ExpectRows bool   `yaml:"expect_rows,omitempty"`
	Query      string `yaml:"query"`
}

type Report struct {
	Name     string   `yaml:"name"`
	Tags     []string `yaml:"tags"`
	Queries  []Query  `yaml:"queries"`
	Hint     string   `yaml:"hint"`
	FilePath string   `yaml:"-"`
}

func NewGraphReportCommand(name string) *GraphReportCommand {
	c := &GraphReportCommand{
		Command: command.Command{
			Name:    name,
			FlagSet: flag.NewFlagSet(name, flag.ExitOnError),
		},
	}
	c.FlagSet().IntVar(&c.logLevel, "log", 4, "Loglevel")
	c.FlagSet().Func("tag", "run all reports with specified tag", func(val string) error {
		c.optTags = append(c.optTags, val)
		return nil
	})
	c.FlagSet().Func("name", "run report with specified report name(s)", func(val string) error {
		names := strings.Split(val, ";")
		for _, n := range names {
			n = strings.TrimSpace(n)
			if n != "" {
				c.optNames = append(c.optNames, n)
			}
		}
		return nil
	})
	c.FlagSet().StringVar(&c.optDb, "db", "bolt://localhost:7687", "database connection string")
	c.FlagSet().StringVar(&c.optReport, "reports", "", "run all reports form the specified reports folder")
	c.FlagSet().BoolVar(&c.optList, "list", false, "list all available reports and their tags")
	c.FlagSet().BoolVar(&c.optListTags, "list-tags", false, "list all available tags from reports")
	c.FlagSet().BoolVar(&c.optListAll, "list-all", false, "list all available report details in tabular format")
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
	if !c.optList && !c.optListTags && !c.optListAll && c.FlagSet().NArg() != 1 {
		return fmt.Errorf("Specify simulation path OR Use --list option")
	}
	return nil
}

func (c *GraphReportCommand) Run() error {
	slog.SetDefault(log.NewLogger(c.logLevel))
	slog.Info("Connect to graph", "db", c.optDb)
	ctx := context.Background()
	driver, err := graph.Driver(c.optDb)
	if err != nil {
		return err
	}
	ctx = context.WithValue(ctx, "driver", driver)
	defer graph.Close(ctx)

	session, err := graph.Session(ctx)
	if err != nil {
		return err
	}
	defer session.Close(ctx)

	if !c.optList && !c.optListTags && !c.optListAll {
		simPath := c.FlagSet().Arg(0)
		var yamlFiles []string
		err = filepath.Walk(simPath, func(p string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if !info.IsDir() && (strings.HasSuffix(p, ".yaml") || strings.HasSuffix(p, ".yml")) {
				yamlFiles = append(yamlFiles, p)
			}
			return nil
		})
		if err != nil {
			return err
		}

		fmt.Println()
		fmt.Println("=== Files ===================================================================")
		for _, fullPath := range yamlFiles {
			index := strings.Index(fullPath, "sim")
			if index != -1 {
				relPath := fullPath[index:]
				fmt.Println(relPath)
			}
		}

		output := os.Stdout
		f, _ := os.Open(os.DevNull)
		os.Stdout = f

		// Import simulation configuration files.
		handler := &kind.YamlKindHandler{}
		for _, yamlFile := range yamlFiles {
			handler.Import(ctx, yamlFile, handler.Detect(yamlFile))
		}
		(&GraphImportCommand{}).createRelationships(ctx, session)

		os.Stdout = output
	}

	var reportPaths []string

	if c.optReport != "" {
		info, err := os.Stat(c.optReport)
		if err != nil {
			return fmt.Errorf("invalid report path: %w", err)
		}

		if info.IsDir() {
			entries, err := os.ReadDir(c.optReport)
			if err != nil {
				return fmt.Errorf("failed to read report directory: %w", err)
			}
			for _, entry := range entries {
				if !entry.IsDir() && (strings.HasSuffix(entry.Name(), ".yaml") || strings.HasSuffix(entry.Name(), ".yml")) {
					reportPaths = append(reportPaths, filepath.Join(c.optReport, entry.Name()))
				}
			}
		} else {
			reportPaths = []string{c.optReport}
		}
	} else {
		// Default directory.
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("failed to get home directory: %w", err)
		}
		defaultDir := filepath.Join(homeDir, ".local", "share", "dse-graph", "reports")
		entries, err := os.ReadDir(defaultDir)
		if err != nil {
			return fmt.Errorf("failed to read default report directory: %w", err)
		}
		for _, entry := range entries {
			if !entry.IsDir() && (strings.HasSuffix(entry.Name(), ".yaml") || strings.HasSuffix(entry.Name(), ".yml")) {
				reportPaths = append(reportPaths, filepath.Join(defaultDir, entry.Name()))
			}
		}
	}

	var (
		reports       []Report
		tagSet        = make(map[string]struct{})
		totalReports  int
		passedReports int
		failedReports int
		failedList    []string
		passedList    []string
	)

	// Allow ; seperated report names.
	var flattened []string
	for _, val := range c.optNames {
		for _, part := range strings.Split(val, ";") {
			part = strings.TrimSpace(part)
			if part != "" {
				flattened = append(flattened, part)
			}
		}
	}
	c.optNames = flattened

	for _, reportPath := range reportPaths {
		name := filepath.Base(reportPath)

		f, err := os.Open(reportPath)
		if err != nil {
			slog.Warn("Failed to open report file", "file", name, "err", err)
			continue
		}

		decoder := yaml.NewDecoder(f)

		for {
			var r Report
			err := decoder.Decode(&r)
			if err != nil {
				if errors.Is(err, io.EOF) {
					break
				}
				slog.Warn("Failed to unmarshal report", "file", name, "err", err)
				break
			}
			r.FilePath = reportPath

			for _, tag := range r.Tags {
				tagSet[tag] = struct{}{}
			}

			if (len(c.optTags) > 0 && !hasTag(r.Tags, c.optTags)) || (len(c.optNames) > 0 && !hasName(r.Name, c.optNames)) {
				continue
			}
			reports = append(reports, r)
		}

		f.Close()
	}

	// List all details.
	if c.optListAll {
		fmt.Println("\nListing all report details:")
		t := table.NewWriter()
		t.SetOutputMirror(os.Stdout)

		t.AppendHeader(table.Row{"Name", "Tags", "File", "Hint"})

		for _, r := range reports {
			hint := r.Hint
			if hint == "" {
				hint = "<no hint>"
			}
			t.AppendRow(table.Row{
				r.Name,
				strings.Join(r.Tags, ", "),
				r.FilePath,
				hint,
			})
		}

		t.Render()
		return nil
	}

	// List tags.
	if c.optListTags {
		fmt.Println("\nListing all report tags:")
		tags := make([]string, 0, len(tagSet))
		for tag := range tagSet {
			tags = append(tags, tag)
		}
		sort.Strings(tags)
		for _, tag := range tags {
			fmt.Println(tag)
		}
		return nil
	}

	// List reports.
	if c.optList {
		fmt.Println("\nListing all report names and tags:")
		sort.SliceStable(reports, func(i, j int) bool {
			return reports[i].Name < reports[j].Name
		})
		for _, r := range reports {
			fmt.Printf("%s [%s]\n", r.Name, strings.Join(r.Tags, ", "))
		}
		return nil
	}

	// Run the reports.
	for _, r := range reports {
		totalReports++
		if err := c.runReport(ctx, session, r.FilePath, r); err != nil {
			failedReports++
			failedList = append(failedList, r.Name)
		} else {
			passedReports++
			passedList = append(passedList, r.Name)
		}
	}

	// Print summary.
	fmt.Println()
	fmt.Println("=== Summary ===================================================================")
	for _, pass := range passedList {
		fmt.Printf("[PASS] %s\n", pass)
	}
	for _, fail := range failedList {
		fmt.Printf("[FAIL] %s\n", fail)
	}
	fmt.Printf("Ran %d Reports | Passed: %d | Failed: %d\n", totalReports, passedReports, failedReports)

	if failedReports > 0 {
		os.Exit(1)
	}

	return nil
}

func (c *GraphReportCommand) runReport(ctx context.Context, session neo4j.SessionWithContext, fileOrFolder string, report Report) error {
	fmt.Println()
	fmt.Println("=== Report ===================================================================")
	fmt.Println("Name:", report.Name)
	fmt.Println("Path:", fileOrFolder)
	fmt.Println("Version: 0.0.0")
	fmt.Printf("Date: %s\n", time.Now().Format("2006-01-02 15:04:05"))

	// Check if there are queries.
	if len(report.Queries) == 0 {
		slog.Error("No queries found in report YAML", "file", fileOrFolder)
	}

	var failed bool
	for _, q := range report.Queries {
		fmt.Println("Query:", q.Name)
		fmt.Println("Cypher:")

		lines := strings.Split(q.Query, "\n")
		if len(lines) > 0 && strings.TrimSpace(lines[len(lines)-1]) == "" {
			lines = lines[:len(lines)-1]
		}
		for _, line := range lines {
			fmt.Printf("    %s\n", line)
		}

		result, err := session.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
			queryResult, err := tx.Run(ctx, q.Query, nil)
			if err != nil {
				return nil, err
			}
			return queryResult.Collect(ctx)
		})
		if err != nil {
			slog.Error("Failed to execute query", "error", err)
			failed = true
			continue
		}

		records, ok := result.([]*neo4j.Record)
		if !ok {
			slog.Error("Unexpected result type from query execution")
			failed = true
			continue
		}

		fmt.Println("Results:")
		printTable(records)

		failedQuery := false

		if q.Evaluate && q.ExpectRows {
			for _, record := range records {
				resultValue, _ := record.Get("result")
				strVal, ok := resultValue.(string)
				if !ok || strVal != "PASS" {
					failedQuery = true
					break
				}
			}
		} else if q.ExpectRows {
			if len(records) == 0 {
				failedQuery = true
			}
		} else {
			if len(records) != 0 {
				failedQuery = true
			}
		}

		// If we are NOT skipping report status logs
		if !(q.ExpectRows && !q.Evaluate) {
			if q.ExpectRows && len(records) == 0 {
				fmt.Println("No records found")
			}

			if failedQuery {
				failed = true
				fmt.Println("Evaluation: Report Failed")
				if report.Hint != "" {
					fmt.Println("Hint:", report.Hint)
				}
			} else {
				fmt.Println("Evaluation: Report Passed")
			}
		}
	}

	if failed {
		return fmt.Errorf("One or more queries failed")
	}
	return nil
}

func hasTag(reportTags, checkTags []string) bool {
	for _, c := range checkTags {
		for _, t := range reportTags {
			if t == c {
				return true
			}
		}
	}
	return false
}

func hasName(name string, optNames []string) bool {
	name = strings.TrimSpace(name)
	for _, opt := range optNames {
		if name == strings.TrimSpace(opt) {
			return true
		}
	}
	return false
}

// Print results in a simple table.
func printTable(records []*neo4j.Record) {
	if len(records) == 0 {
		return
	}

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

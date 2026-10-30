package graph

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"slices"
	"sort"
	"strings"
	"time"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
	"gopkg.in/yaml.v3"

	"github.com/boschglobal/dse.sdp/graph/internal/pkg/file/kind"
)

var errReportsFailed = errors.New("one or more reports failed")

type GraphReportCommand struct {
	base
	optTags     []string
	optNames    []string
	optReport   string
	optList     bool
	optListTags bool
	optListAll  bool
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
	c := &GraphReportCommand{}
	c.init(name)
	c.FlagSet().Func("tag", "run all reports with specified tag", func(val string) error {
		c.optTags = append(c.optTags, val)
		return nil
	})
	// Report names may be given as a ; separated list.
	c.FlagSet().Func("name", "run report with specified report name(s)", func(val string) error {
		for _, n := range strings.Split(val, ";") {
			if n = strings.TrimSpace(n); n != "" {
				c.optNames = append(c.optNames, n)
			}
		}
		return nil
	})
	c.FlagSet().StringVar(&c.optReport, "reports", "", "run all reports form the specified reports folder")
	c.FlagSet().BoolVar(&c.optList, "list", false, "list all available reports and their tags")
	c.FlagSet().BoolVar(&c.optListTags, "list-tags", false, "list all available tags from reports")
	c.FlagSet().BoolVar(&c.optListAll, "list-all", false, "list all available report details in tabular format")
	return c
}

func (c *GraphReportCommand) Parse(args []string) error {
	if err := c.base.Parse(args); err != nil {
		return err
	}
	if !c.listing() && c.FlagSet().NArg() != 1 {
		return errors.New("specify simulation path OR use --list option")
	}
	return nil
}

func (c *GraphReportCommand) listing() bool {
	return c.optList || c.optListTags || c.optListAll
}

func (c *GraphReportCommand) Run() error {
	ctx, session, closeFn, err := c.connect()
	if err != nil {
		return err
	}
	defer closeFn()

	if !c.listing() {
		if err := importSimulation(ctx, session, c.FlagSet().Arg(0)); err != nil {
			return err
		}
	}

	reportPaths, err := c.reportPaths()
	if err != nil {
		return err
	}
	reports, tags := c.loadReports(reportPaths)

	switch {
	case c.optListAll:
		listAll(reports)
	case c.optListTags:
		listTags(tags)
	case c.optList:
		listReports(reports)
	default:
		return c.runReports(ctx, session, reports)
	}
	return nil
}

func importSimulation(ctx context.Context, session neo4j.SessionWithContext, simPath string) error {
	yamlFiles, err := findYAMLFiles(simPath)
	if err != nil {
		return err
	}

	fmt.Println()
	fmt.Println("=== Files ===================================================================")
	for _, file := range yamlFiles {
		fmt.Println(file)
	}

	importErr := importKindFiles(ctx, &kind.YamlKindHandler{Out: io.Discard}, yamlFiles)
	createRelationships(ctx, session)
	if importErr != nil {
		return fmt.Errorf("import simulation: %w", importErr)
	}
	return nil
}

func (c *GraphReportCommand) reportPaths() ([]string, error) {
	if c.optReport == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("failed to get home directory: %w", err)
		}
		dir := filepath.Join(homeDir, ".local", "share", "dse-graph", "reports")
		paths, err := yamlFilesInDir(dir)
		if err != nil {
			return nil, fmt.Errorf("failed to read default report directory: %w", err)
		}
		return paths, nil
	}

	info, err := os.Stat(c.optReport)
	if err != nil {
		return nil, fmt.Errorf("invalid report path: %w", err)
	}
	if !info.IsDir() {
		return []string{c.optReport}, nil
	}
	paths, err := yamlFilesInDir(c.optReport)
	if err != nil {
		return nil, fmt.Errorf("failed to read report directory: %w", err)
	}
	return paths, nil
}

func yamlFilesInDir(dir string) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	var paths []string
	for _, entry := range entries {
		if !entry.IsDir() && isYAMLFile(entry.Name()) {
			paths = append(paths, filepath.Join(dir, entry.Name()))
		}
	}
	return paths, nil
}

func (c *GraphReportCommand) loadReports(paths []string) ([]Report, map[string]struct{}) {
	var reports []Report
	tags := make(map[string]struct{})
	for _, path := range paths {
		for _, r := range decodeReports(path) {
			for _, tag := range r.Tags {
				tags[tag] = struct{}{}
			}
			if c.selected(r) {
				reports = append(reports, r)
			}
		}
	}
	return reports, tags
}

func (c *GraphReportCommand) selected(r Report) bool {
	if len(c.optTags) > 0 && !hasTag(r.Tags, c.optTags) {
		return false
	}
	if len(c.optNames) > 0 && !hasName(r.Name, c.optNames) {
		return false
	}
	return true
}

func decodeReports(path string) []Report {
	name := filepath.Base(path)
	f, err := os.Open(path)
	if err != nil {
		slog.Warn("Failed to open report file", "file", name, "err", err)
		return nil
	}
	defer f.Close()

	var reports []Report
	decoder := yaml.NewDecoder(f)
	for {
		var r Report
		if err := decoder.Decode(&r); err != nil {
			if !errors.Is(err, io.EOF) {
				slog.Warn("Failed to unmarshal report", "file", name, "err", err)
			}
			return reports
		}
		r.FilePath = path
		reports = append(reports, r)
	}
}

func listAll(reports []Report) {
	fmt.Println("\nListing all report details:")
	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.AppendHeader(table.Row{"Name", "Tags", "File", "Hint"})
	for _, r := range reports {
		hint := r.Hint
		if hint == "" {
			hint = "<no hint>"
		}
		t.AppendRow(table.Row{r.Name, strings.Join(r.Tags, ", "), r.FilePath, hint})
	}
	t.Render()
}

func listTags(tagSet map[string]struct{}) {
	fmt.Println("\nListing all report tags:")
	tags := make([]string, 0, len(tagSet))
	for tag := range tagSet {
		tags = append(tags, tag)
	}
	sort.Strings(tags)
	for _, tag := range tags {
		fmt.Println(tag)
	}
}

func listReports(reports []Report) {
	fmt.Println("\nListing all report names and tags:")
	sort.SliceStable(reports, func(i, j int) bool {
		return reports[i].Name < reports[j].Name
	})
	for _, r := range reports {
		fmt.Printf("%s [%s]\n", r.Name, strings.Join(r.Tags, ", "))
	}
}

func (c *GraphReportCommand) runReports(ctx context.Context, session neo4j.SessionWithContext, reports []Report) error {
	var passed, failed []string
	for _, r := range reports {
		if err := c.runReport(ctx, session, r.FilePath, r); err != nil {
			failed = append(failed, r.Name)
		} else {
			passed = append(passed, r.Name)
		}
	}

	fmt.Println()
	fmt.Println("=== Summary ===================================================================")
	for _, name := range passed {
		fmt.Printf("[PASS] %s\n", name)
	}
	for _, name := range failed {
		fmt.Printf("[FAIL] %s\n", name)
	}
	fmt.Printf("Ran %d Reports | Passed: %d | Failed: %d\n", len(reports), len(passed), len(failed))

	if len(failed) > 0 {
		return errReportsFailed
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

	if len(report.Queries) == 0 {
		slog.Error("No queries found in report YAML", "file", fileOrFolder)
	}

	failed := false
	for _, q := range report.Queries {
		if !runQuery(ctx, session, q, report.Hint) {
			failed = true
		}
	}
	if failed {
		return errors.New("one or more queries failed")
	}
	return nil
}

func runQuery(ctx context.Context, session neo4j.SessionWithContext, q Query, hint string) bool {
	fmt.Println("Query:", q.Name)
	fmt.Println("Cypher:")
	for _, line := range strings.Split(strings.TrimRight(q.Query, " \t\r\n"), "\n") {
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
		return false
	}
	records, ok := result.([]*neo4j.Record)
	if !ok {
		slog.Error("Unexpected result type from query execution")
		return false
	}

	fmt.Println("Results:")
	printTable(records)

	// Queries which expect rows without evaluation report no status.
	if q.ExpectRows && !q.Evaluate {
		return true
	}
	if q.ExpectRows && len(records) == 0 {
		fmt.Println("No records found")
	}
	if isQueryFailed(q, records) {
		fmt.Println("Evaluation: Report Failed")
		if hint != "" {
			fmt.Println("Hint:", hint)
		}
		return false
	}
	fmt.Println("Evaluation: Report Passed")
	return true
}

func isQueryFailed(q Query, records []*neo4j.Record) bool {
	switch {
	case q.ExpectRows && q.Evaluate:
		return slices.ContainsFunc(records, func(r *neo4j.Record) bool {
			v, _ := recordValue[string](r, "result")
			return v != "PASS"
		})
	case q.ExpectRows:
		return len(records) == 0
	default:
		return len(records) != 0
	}
}

func hasTag(reportTags, checkTags []string) bool {
	return slices.ContainsFunc(checkTags, func(c string) bool {
		return slices.Contains(reportTags, c)
	})
}

func hasName(name string, optNames []string) bool {
	name = strings.TrimSpace(name)
	return slices.ContainsFunc(optNames, func(opt string) bool {
		return name == strings.TrimSpace(opt)
	})
}

// Print results in a simple table.
func printTable(records []*neo4j.Record) {
	if len(records) == 0 {
		return
	}

	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)

	headers := records[0].Keys
	header := make(table.Row, len(headers))
	for i, col := range headers {
		header[i] = col
	}
	t.AppendHeader(header)

	for _, record := range records {
		row := make(table.Row, len(headers))
		for i, col := range headers {
			row[i], _ = record.Get(col)
		}
		t.AppendRow(row)
	}
	t.Render()
}

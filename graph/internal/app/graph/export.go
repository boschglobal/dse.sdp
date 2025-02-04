package graph

import (
	"context"
	"flag"
	"log/slog"
	"os"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"

	"github.boschdevcloud.com/fsil/fsil.go/command"

	"github.com/boschglobal/dse.sdp/graph/internal/pkg/graph"
)

type GraphExportCommand struct {
	command.Command
	optDb     string
}

func NewGraphExportCommand(name string) *GraphExportCommand {
	c := &GraphExportCommand{
		Command: command.Command{
			Name:    name,
			FlagSet: flag.NewFlagSet(name, flag.ExitOnError),
		},
	}
	c.FlagSet().StringVar(&c.optDb, "db", "bolt://localhost:7687", "database connection string")
	return c
}

// CommandRunner interface functions.
func (c GraphExportCommand) Name() string {
	return c.Command.Name
}

func (c GraphExportCommand) FlagSet() *flag.FlagSet {
	return c.Command.FlagSet
}

func (c *GraphExportCommand) Parse(args []string) error {
	return c.FlagSet().Parse(args)
}

func (c *GraphExportCommand) Run() error {
	slog.Info("Connect to graph", "db", c.optDb)
	ctx := context.Background()
	driver, err := graph.Driver(c.optDb)
	if err != nil {
		slog.Info("Graph driver error", "error", err)
		return err
	}
	ctx = context.WithValue(ctx, "driver", driver)
	defer graph.Close(ctx)

	session, err := graph.Session(ctx)
	if err != nil {
		slog.Info("Graph session error", "err", err)
		return err
	}
	defer session.Close(ctx)

	args := c.FlagSet().Args()
	if len(args) == 0 {
		slog.Info("Usage: graph export <file>")
		return nil
	}
	file := args[0]
	return c.export(ctx, file)

	return nil
}

func (c *GraphExportCommand) export(ctx context.Context, file string) error {
	slog.Info("Graph Export", "file", file)
	session, _ := graph.Session(ctx)
	defer session.Close(ctx)

	r, err := graph.QueryRecord(ctx, session,
		"CALL export_util.cypher_all(\"\", {stream: true}) YIELD data RETURN data",
		map[string]any{})
	if err != nil {
		return err
	}
	data, _, err := neo4j.GetRecordValue[string](r, "data")
	slog.Info("Graph Export: write", "file", file, "data", data)
	f, err := os.Create(file)
	if err != nil {
		return err
	}
	defer f.Close()
	f.WriteString(data + "\n")
	return nil
}

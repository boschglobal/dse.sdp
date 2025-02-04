package graph

import (
	"context"
	"flag"
	"log/slog"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"

	"github.boschdevcloud.com/fsil/fsil.go/command"

	"github.com/boschglobal/dse.sdp/graph/internal/pkg/graph"
	"github.com/boschglobal/dse.sdp/graph/internal/pkg/file/kind"
)

type GraphImportCommand struct {
	command.Command
	optImport string
	optDb     string
}

func NewGraphImportCommand(name string) *GraphImportCommand {
	c := &GraphImportCommand{
		Command: command.Command{
			Name:    name,
			FlagSet: flag.NewFlagSet(name, flag.ExitOnError),
		},
	}
	c.FlagSet().StringVar(&c.optImport, "import", "", "import files to the database")
	c.FlagSet().StringVar(&c.optDb, "db", "bolt://localhost:7687", "database connection string")
	return c
}

// CommandRunner interface functions.
func (c GraphImportCommand) Name() string {
	return c.Command.Name
}

func (c GraphImportCommand) FlagSet() *flag.FlagSet {
	return c.Command.FlagSet
}

func (c *GraphImportCommand) Parse(args []string) error {
	return c.FlagSet().Parse(args)
}

func (c *GraphImportCommand) Run() error {
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

	args := c.FlagSet().Args() // Get positional arguments
	if len(args) == 0 {
		slog.Info("Usage: graph import <yaml-file>")
		return nil
	}
	file := args[0]
	c.importFiles(ctx, file)

	return nil
}

func (c *GraphImportCommand) matchNode(ctx context.Context, session neo4j.SessionWithContext) {
	match_instance := `
	MATCH (ast_mi:Ast:ModelInst), (sim_mi:Sim:ModelInst)
    WHERE ast_mi.model_name = sim_mi.name
    MERGE (sim_mi)-[:Represents]->(ast_mi)
    `
	_, err := graph.Query(ctx, session, match_instance, nil)
	if err != nil {
		slog.Info("Failed to create relationship", "error", err)
	}

	match_channel := `
	MATCH (ast_ch:Ast:SimulationChannel), (sim_ch:Sim:Channel)
	WHERE ast_ch.channel_name = sim_ch.name
	MERGE (sim_ch)-[:Represents]->(ast_ch)
	`
	_, err = graph.Query(ctx, session, match_channel, nil)
	if err != nil {
		slog.Info("Failed to create relationship", "error", err)
	}
}

func (c *GraphImportCommand) importFiles(ctx context.Context, file string) {
    if file == "" {
        slog.Info("Usage: import <yaml-file>")
    }
    // Initialize the YAML handler.
    handler := &kind.YamlKindHandler{}
    data := handler.Detect(file)

    // Connect to the database.
    driver, err := graph.Driver(c.optDb)
    if err != nil {
        slog.Info("Failed to connect to graph database", "error", err)
    }
    ctx = context.WithValue(ctx, "driver", driver)
    defer graph.Close(ctx)

    // Import the data into the database.
	handler.Import(ctx, file, data)

}

package graph

import (
	"context"
	"flag"
	"log/slog"
	"strings"

	"github.boschdevcloud.com/fsil/fsil.go/command"

	"github.com/boschglobal/dse.sdp/graph/internal/pkg/graph"
)

type GraphDropCommand struct {
	command.Command
	optDb   string
	optAll  bool
}

func NewGraphDropCommand(name string) *GraphDropCommand {
	c := &GraphDropCommand{
		Command: command.Command{
			Name:    name,
			FlagSet: flag.NewFlagSet(name, flag.ExitOnError),
		},
	}
	c.FlagSet().BoolVar(&c.optAll, "all", false, "drop nodes based on label from the database (Usage: drop ast, drop sim, drop --all)")
	c.FlagSet().StringVar(&c.optDb, "db", "bolt://localhost:7687", "database connection string")
	return c
}

// CommandRunner interface functions.
func (c GraphDropCommand) Name() string {
	return c.Command.Name
}

func (c GraphDropCommand) FlagSet() *flag.FlagSet {
	return c.Command.FlagSet
}

func (c *GraphDropCommand) Parse(args []string) error {
	return c.FlagSet().Parse(args)
}

func (c *GraphDropCommand) Run() error {
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

	if c.optAll {
		c.runDrop(ctx, "--all")
	} else if len(args) > 0 {
		option := strings.ToLower(args[0]) // Handle ast/sim as case-insensitive
		if option == "ast" || option == "sim" {
			c.runDrop(ctx, option)
		} else {
			slog.Info("Invalid usage. Use 'graph drop ast', 'graph drop sim', or 'graph drop --all'")
			return nil
		}
	} else {
		slog.Info("Invalid usage. Use 'graph drop ast', 'graph drop sim', or 'graph drop --all'")
		return nil
	}

	return nil
}

// Internal implementation.
func (c *GraphDropCommand) runDrop(ctx context.Context, option string) {
	graph.Drop(ctx, option)
}

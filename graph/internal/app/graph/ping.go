package graph

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/boschglobal/dse.clib/extra/go/command"

	"github.com/boschglobal/dse.sdp/graph/internal/pkg/graph"
)

type GraphPingCommand struct {
	command.Command
	optRetries int
	optDb      string
}

func NewGraphPingCommand(name string) *GraphPingCommand {
	c := &GraphPingCommand{
		Command: command.Command{
			Name:    name,
			FlagSet: flag.NewFlagSet(name, flag.ExitOnError),
		},
	}
	c.FlagSet().IntVar(&c.optRetries, "retry", 30, "number of retries")
	c.FlagSet().StringVar(&c.optDb, "db", "bolt://localhost:7687", "database connection string")
	return c
}

// CommandRunner interface functions.
func (c GraphPingCommand) Name() string {
	return c.Command.Name
}

func (c GraphPingCommand) FlagSet() *flag.FlagSet {
	return c.Command.FlagSet
}

func (c *GraphPingCommand) Parse(args []string) error {
	return c.FlagSet().Parse(args)
}

func (c *GraphPingCommand) Run() error {
	slog.Debug("Connect to graph", "db", c.optDb)

	for i := 0; i < c.optRetries; i++ {
		if i > 0 {
			time.Sleep(1 * time.Second)
		}
		ctx := context.Background()

		// Driver.
		driver, err := graph.Driver(c.optDb)
		if err != nil {
			fmt.Fprintf(os.Stderr, "ping: failed (%v)\n", err)
			continue
		}
		ctx = context.WithValue(ctx, "driver", driver)
		defer graph.Close(ctx)

		// Session.
		session, err := graph.Session(ctx)
		if err != nil {
			slog.Info("Graph session error", "err", err)
			return err
		}
		defer session.Close(ctx)

		// Database.
		_, err = session.Run(ctx, "return 1", nil)
		if err != nil {
			fmt.Fprintf(os.Stderr, "ping: failed (%v)\n", err)
			continue
		}
		fmt.Println("ping: OK")
		return nil
	}

	return fmt.Errorf("No connection established!")
}

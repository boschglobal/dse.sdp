package graph

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/boschglobal/dse.sdp/graph/internal/pkg/graph"
)

type GraphPingCommand struct {
	base
	optRetries int
}

func NewGraphPingCommand(name string) *GraphPingCommand {
	c := &GraphPingCommand{}
	c.init(name)
	c.FlagSet().IntVar(&c.optRetries, "retry", 30, "number of retries")
	return c
}

func (c *GraphPingCommand) Run() error {
	c.setupLogging()
	slog.Debug("Connect to graph", "db", c.optDb)

	for i := 0; i < c.optRetries; i++ {
		if i > 0 {
			time.Sleep(1 * time.Second)
		}
		if err := c.ping(context.Background()); err != nil {
			fmt.Fprintf(os.Stderr, "ping: failed (%v)\n", err)
			continue
		}
		fmt.Println("ping: OK")
		return nil
	}

	return errors.New("no connection established")
}

func (c *GraphPingCommand) ping(ctx context.Context) error {
	ctx, session, closeFn, err := graph.Connect(ctx, c.optDb)
	if err != nil {
		return err
	}
	defer closeFn()

	result, err := session.Run(ctx, "return 1", nil)
	if err != nil {
		return err
	}
	_, err = result.Consume(ctx)
	return err
}

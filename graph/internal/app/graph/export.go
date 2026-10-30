package graph

import (
	"context"
	"errors"
	"log/slog"
	"os"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"

	"github.com/boschglobal/dse.sdp/graph/internal/pkg/graph"
)

type GraphExportCommand struct {
	base
}

func NewGraphExportCommand(name string) *GraphExportCommand {
	c := &GraphExportCommand{}
	c.init(name)
	return c
}

func (c *GraphExportCommand) Run() error {
	c.setupLogging()
	args := c.FlagSet().Args()
	if len(args) == 0 {
		return errors.New("usage: graph export <file>")
	}

	ctx, session, closeFn, err := c.connect()
	if err != nil {
		return err
	}
	defer closeFn()

	return export(ctx, session, args[0])
}

func export(ctx context.Context, session neo4j.SessionWithContext, file string) error {
	slog.Info("Graph Export", "file", file)
	r, err := graph.QueryRecord(ctx, session,
		"CALL export_util.cypher_all(\"\", {stream: true}) YIELD data RETURN data",
		map[string]any{})
	if err != nil {
		return err
	}
	if r == nil {
		return errors.New("graph export returned no data")
	}
	data, _, err := neo4j.GetRecordValue[string](r, "data")
	if err != nil {
		return err
	}
	slog.Info("Graph Export: write", "file", file, "bytes", len(data))
	return os.WriteFile(file, []byte(data+"\n"), 0o644)
}

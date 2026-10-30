package graph

import (
	"context"
	"flag"
	"log/slog"
	"os"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"

	"github.com/boschglobal/dse.clib/extra/go/command"
	"github.com/boschglobal/dse.clib/extra/go/command/log"

	"github.com/boschglobal/dse.sdp/graph/internal/pkg/graph"
)

// Base type of CommandRunner functions.
type base struct {
	command.Command
	logLevel int
	optDb    string
}

// Init, call after creating the command type/value (i.e. in NewCommand), and before using.
func (b *base) init(name string) {
	b.Command = command.Command{
		Name:    name,
		FlagSet: flag.NewFlagSet(name, flag.ExitOnError),
	}
	b.Command.FlagSet.IntVar(&b.logLevel, "log", 4, "log level: 0=trace, 1=debug, 2=info, 3=warn, 4=error")
	b.Command.FlagSet.StringVar(&b.optDb, "db", "bolt://localhost:7687", "database connection string")
}

func (b *base) Name() string {
	return b.Command.Name
}

func (b *base) FlagSet() *flag.FlagSet {
	return b.Command.FlagSet
}

func (b *base) Parse(args []string) error {
	return b.FlagSet().Parse(args)
}

func (b *base) setupLogging() {
	slog.SetDefault(newLogger(b.logLevel))
}

func newLogger(level int) *slog.Logger {
	levels := []slog.Level{graph.LevelTrace, slog.LevelDebug, slog.LevelInfo, slog.LevelWarn, slog.LevelError}
	lvl := slog.LevelWarn
	if level >= 0 && level < len(levels) {
		lvl = levels[level]
	}
	return slog.New(log.NewPlainLogHandler(os.Stderr, log.PlainLogHandlerOptions{
		SlogOpts: slog.HandlerOptions{Level: lvl},
	}))
}

func (b *base) connect() (ctx context.Context, session neo4j.SessionWithContext, closeFn func(), err error) {
	b.setupLogging()
	slog.Info("Connect to graph", "db", b.optDb)
	return graph.Connect(context.Background(), b.optDb)
}

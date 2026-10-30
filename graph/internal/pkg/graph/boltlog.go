package graph

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j/log"
)

// LevelTrace is the slog level for the complete Bolt traffic, below slog.LevelDebug.
const LevelTrace = slog.Level(-8)

// boltLogger logs the Bolt messages of the driver through slog: at debug level
// each query and the summary of its response, at trace level all messages
// including the parameters and every record.
type boltLogger struct{}

var _ log.BoltLogger = boltLogger{}

func (boltLogger) LogClientMessage(id, msg string, args ...any) {
	logBolt("client", id, msg, args)
}

func (boltLogger) LogServerMessage(id, msg string, args ...any) {
	logBolt("server", id, msg, args)
}

// boltLevel returns the level at which a Bolt message is logged.
func boltLevel(msg string) slog.Level {
	switch name, _, _ := strings.Cut(msg, " "); name {
	case "RUN", "BEGIN", "COMMIT", "ROLLBACK", "SUCCESS", "FAILURE", "IGNORED":
		return slog.LevelDebug
	}
	return LevelTrace
}

func logBolt(direction, id, msg string, args []any) {
	ctx := context.Background()
	logger := slog.Default()
	level := boltLevel(msg)
	if !logger.Enabled(ctx, level) {
		return
	}

	text := fmt.Sprintf(msg, args...)
	// Parameters may hold sensitive values, show only the query unless tracing.
	if strings.HasPrefix(msg, "RUN") && !logger.Enabled(ctx, LevelTrace) && len(args) > 0 {
		text = fmt.Sprintf("RUN %q", args[0])
	}
	logger.Log(ctx, level, "Graph "+direction, "conn", id, "msg", text)
}

// driverLogger logs the internal events of the driver (pool, routing) through slog.
type driverLogger struct{}

var _ log.Logger = driverLogger{}

func (driverLogger) Error(name, id string, err error) {
	slog.Debug("Graph driver", "component", name, "id", id, "err", err)
}

func (driverLogger) Warnf(name, id, msg string, args ...any) {
	if slog.Default().Enabled(context.Background(), slog.LevelWarn) {
		slog.Warn("Graph driver", "component", name, "id", id, "msg", fmt.Sprintf(msg, args...))
	}
}

func (driverLogger) Infof(name, id, msg string, args ...any) {
	logDriver(name, id, msg, args)
}

func (driverLogger) Debugf(name, id, msg string, args ...any) {
	logDriver(name, id, msg, args)
}

func logDriver(name, id, msg string, args []any) {
	ctx := context.Background()
	if logger := slog.Default(); logger.Enabled(ctx, LevelTrace) {
		logger.Log(ctx, LevelTrace, "Graph driver", "component", name, "id", id, "msg", fmt.Sprintf(msg, args...))
	}
}

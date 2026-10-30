package graph

import (
	"bytes"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
)

// captureLogs makes slog write to the returned buffer at the given level.
func captureLogs(t *testing.T, level slog.Level) *bytes.Buffer {
	t.Helper()
	var buf bytes.Buffer
	previous := slog.Default()
	slog.SetDefault(slog.New(slog.NewTextHandler(&buf, &slog.HandlerOptions{Level: level})))
	t.Cleanup(func() { slog.SetDefault(previous) })
	return &buf
}

func logExchange() {
	var l boltLogger
	l.LogClientMessage("c1", "RUN %q %s %s", "MATCH (n) WHERE n.k = $k RETURN n", "{k: secret}", "{}")
	l.LogClientMessage("c1", "PULL %s", "{n: 1000}")
	l.LogServerMessage("c1", "SUCCESS %s", "{fields: [n]}")
	l.LogServerMessage("c1", "RECORD %s", "[node-1]")
	l.LogServerMessage("c1", "SUCCESS %s", "{t_last: 3}")
}

func TestBoltLoggerInfoIsSilent(t *testing.T) {
	buf := captureLogs(t, slog.LevelInfo)
	logExchange()
	assert.Empty(t, buf.String())
}

func TestBoltLoggerDebugShowsQueryAndSummary(t *testing.T) {
	buf := captureLogs(t, slog.LevelDebug)
	logExchange()

	out := buf.String()
	assert.Contains(t, out, "MATCH (n) WHERE n.k = $k RETURN n")
	assert.Contains(t, out, "fields: [n]")
	assert.Contains(t, out, "t_last: 3")
	assert.NotContains(t, out, "secret", "parameters are only logged when tracing")
	assert.NotContains(t, out, "node-1", "records are only logged when tracing")
	assert.NotContains(t, out, "PULL")
}

func TestBoltLoggerTraceShowsEverything(t *testing.T) {
	buf := captureLogs(t, LevelTrace)
	logExchange()

	out := buf.String()
	assert.Contains(t, out, "secret")
	assert.Contains(t, out, "node-1")
	assert.Contains(t, out, "PULL")
}

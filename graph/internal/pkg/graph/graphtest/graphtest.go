// Package graphtest provides a Memgraph fixture for tests.
//
// Packages call Main from TestMain. The fixture uses the instance named by the
// GRAPH_DB environment variable (e.g. a GitHub Actions service container) or
// else starts a throw-away container with docker. Without either, tests which
// call Connect are skipped, or failed when the CI environment variable is set.
package graphtest

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"

	"github.com/boschglobal/dse.sdp/graph/internal/pkg/graph"
)

const (
	defaultImage = "memgraph/memgraph-mage:3.7.1"
	readyTimeout = 90 * time.Second
)

var (
	uri      string
	startErr error
)

// Main starts the graph instance, runs the tests and releases the instance.
func Main(m *testing.M) {
	os.Exit(run(m))
}

func run(m *testing.M) int {
	var stop func()
	uri, stop, startErr = start()
	if startErr != nil {
		fmt.Fprintln(os.Stderr, "graphtest: graph database unavailable:", startErr)
	}
	code := m.Run()
	if stop != nil {
		stop()
	}
	return code
}

// URI returns the connection string of the fixture database, skipping the
// test when it is unavailable.
func URI(t testing.TB) string {
	t.Helper()
	if uri == "" {
		if os.Getenv("CI") != "" {
			t.Fatalf("graph database unavailable: %v", startErr)
		}
		t.Skipf("graph database unavailable: %v", startErr)
	}
	return uri
}

// Connect returns a context and session for an empty graph. The connection is
// closed when the test completes.
func Connect(t testing.TB) (context.Context, neo4j.SessionWithContext) {
	t.Helper()
	uri := URI(t)
	ctx, session, closeFn, err := graph.Connect(context.Background(), uri)
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	t.Cleanup(closeFn)
	Cypher(t, ctx, session, "MATCH (n) DETACH DELETE n", nil)
	return ctx, session
}

// Cypher runs a query and returns all of its records.
func Cypher(t testing.TB, ctx context.Context, session neo4j.SessionWithContext, query string, params map[string]any) []*neo4j.Record {
	t.Helper()
	records, err := session.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
		result, err := tx.Run(ctx, query, params)
		if err != nil {
			return nil, err
		}
		return result.Collect(ctx)
	})
	if err != nil {
		t.Fatalf("query %q: %v", query, err)
	}
	return records.([]*neo4j.Record)
}

// Count runs a query returning a single integer column named n.
func Count(t testing.TB, ctx context.Context, session neo4j.SessionWithContext, query string, params map[string]any) int64 {
	t.Helper()
	records := Cypher(t, ctx, session, query, params)
	if len(records) != 1 {
		t.Fatalf("query %q: expected 1 record, got %d", query, len(records))
	}
	n, _, err := neo4j.GetRecordValue[int64](records[0], "n")
	if err != nil {
		t.Fatalf("query %q: %v", query, err)
	}
	return n
}

func start() (string, func(), error) {
	if env := os.Getenv("GRAPH_DB"); env != "" {
		return env, nil, waitReady(env)
	}
	return startContainer()
}

func startContainer() (string, func(), error) {
	if _, err := exec.LookPath("docker"); err != nil {
		return "", nil, fmt.Errorf("set GRAPH_DB or install docker: %w", err)
	}
	image := os.Getenv("GRAPH_IMAGE")
	if image == "" {
		image = defaultImage
	}

	out, err := exec.Command("docker", "run", "-d", "--rm", "-p", "127.0.0.1::7687", image).Output()
	if err != nil {
		return "", nil, fmt.Errorf("docker run %s: %w", image, err)
	}
	id := strings.TrimSpace(string(out))
	stop := func() { _ = exec.Command("docker", "kill", id).Run() }

	out, err = exec.Command("docker", "port", id, "7687/tcp").Output()
	if err != nil {
		stop()
		return "", nil, fmt.Errorf("docker port: %w", err)
	}
	// Output is "127.0.0.1:<port>", possibly one line per address family.
	address, _, _ := strings.Cut(strings.TrimSpace(string(out)), "\n")
	boltURI := "bolt://" + strings.TrimSpace(address)

	if err := waitReady(boltURI); err != nil {
		stop()
		return "", nil, err
	}
	return boltURI, stop, nil
}

// waitReady polls the database until it answers a query.
func waitReady(boltURI string) error {
	ctx, cancel := context.WithTimeout(context.Background(), readyTimeout)
	defer cancel()

	var lastErr error
	for ctx.Err() == nil {
		if lastErr = ping(ctx, boltURI); lastErr == nil {
			return nil
		}
		time.Sleep(500 * time.Millisecond)
	}
	return fmt.Errorf("graph database %s not ready: %w", boltURI, lastErr)
}

func ping(ctx context.Context, boltURI string) error {
	driver, err := neo4j.NewDriverWithContext(boltURI, neo4j.BasicAuth("", "", ""))
	if err != nil {
		return err
	}
	defer driver.Close(ctx)
	if err := driver.VerifyConnectivity(ctx); err != nil {
		return err
	}
	session := driver.NewSession(ctx, neo4j.SessionConfig{})
	defer session.Close(ctx)
	result, err := session.Run(ctx, "RETURN 1", nil)
	if err != nil {
		return err
	}
	_, err = result.Consume(ctx)
	return err
}

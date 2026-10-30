package graph

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"sort"
	"strings"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j/config"
)

type driverKey struct{}

// Executor runs write work. A session runs each call in its own transaction;
// the Executor given by WriteTx runs all calls in one transaction.
type Executor interface {
	ExecuteWrite(ctx context.Context, work neo4j.ManagedTransactionWork, configurers ...func(*neo4j.TransactionConfig)) (any, error)
}

type txExecutor struct {
	tx neo4j.ManagedTransaction
}

func (e txExecutor) ExecuteWrite(_ context.Context, work neo4j.ManagedTransactionWork, _ ...func(*neo4j.TransactionConfig)) (any, error) {
	return work(e.tx)
}

// WriteTx runs fn in a single transaction, which is rolled back if fn fails.
// The driver may run fn more than once, so fn must be idempotent.
func WriteTx(ctx context.Context, session neo4j.SessionWithContext, fn func(Executor) error) error {
	_, err := session.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
		return nil, fn(txExecutor{tx})
	})
	return err
}

// authFromEnv reads the credentials from GRAPH_USER and GRAPH_PASSWORD.
func authFromEnv() neo4j.AuthToken {
	return neo4j.BasicAuth(os.Getenv("GRAPH_USER"), os.Getenv("GRAPH_PASSWORD"), "")
}

// Driver creates a driver for the database at dbUri, authenticating with the
// GRAPH_USER and GRAPH_PASSWORD environment variables (none by default).
func Driver(dbUri string) (neo4j.DriverWithContext, error) {
	return neo4j.NewDriverWithContext(dbUri, authFromEnv(), func(c *config.Config) {
		c.Log = driverLogger{}
	})
}

// WithDriver returns a context carrying the driver, for use with Session and Close.
func WithDriver(ctx context.Context, driver neo4j.DriverWithContext) context.Context {
	return context.WithValue(ctx, driverKey{}, driver)
}

func driverFrom(ctx context.Context) (neo4j.DriverWithContext, bool) {
	driver, ok := ctx.Value(driverKey{}).(neo4j.DriverWithContext)
	return driver, ok && driver != nil
}

// Connect creates a driver and a session, returning a context carrying the
// driver. The returned function releases both and must be called by the caller.
func Connect(ctx context.Context, dbUri string) (context.Context, neo4j.SessionWithContext, func(), error) {
	driver, err := Driver(dbUri)
	if err != nil {
		return ctx, nil, nil, fmt.Errorf("graph driver: %w", err)
	}
	ctx = WithDriver(ctx, driver)
	session, err := Session(ctx)
	if err != nil {
		Close(ctx)
		return ctx, nil, nil, fmt.Errorf("graph session: %w", err)
	}
	closeFn := func() {
		session.Close(ctx)
		Close(ctx)
	}
	return ctx, session, closeFn, nil
}

// Close closes the driver held by the context, if any.
func Close(ctx context.Context) {
	driver, ok := driverFrom(ctx)
	if !ok {
		return
	}
	if err := driver.Close(ctx); err != nil {
		slog.Warn("Graph driver close", "err", err)
	}
}

// Session creates a write session using the driver held by the context.
func Session(ctx context.Context) (neo4j.SessionWithContext, error) {
	driver, ok := driverFrom(ctx)
	if !ok {
		return nil, errors.New("no graph driver in context")
	}
	return driver.NewSession(ctx, neo4j.SessionConfig{
		AccessMode: neo4j.AccessModeWrite,
		BoltLogger: boltLogger{},
	}), nil
}

// Drop deletes nodes from the graph; option is one of "ast", "sim" or "--all".
func Drop(ctx context.Context, option string) error {
	var query string
	switch option {
	case "ast":
		query = "MATCH (n:Ast) DETACH DELETE n"
	case "sim":
		query = "MATCH (n:Sim) DETACH DELETE n"
	case "--all":
		query = "MATCH (n) DETACH DELETE n"
	default:
		return fmt.Errorf("invalid drop option %q, use 'ast', 'sim', or '--all'", option)
	}
	session, err := Session(ctx)
	if err != nil {
		return err
	}
	defer session.Close(ctx)
	_, err = session.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
		return tx.Run(ctx, query, nil)
	})
	return err
}

// optionalRecord returns the only record of a result, nil if it has none, or an
// error if it has more than one.
func optionalRecord(ctx context.Context, result neo4j.ResultWithContext) (*neo4j.Record, error) {
	records, err := result.Collect(ctx)
	if err != nil {
		return nil, err
	}
	switch len(records) {
	case 0:
		return nil, nil
	case 1:
		return records[0], nil
	}
	return nil, fmt.Errorf("expected at most 1 record, got %d", len(records))
}

// executeID runs query and returns the "id" column of its single result row.
// If tolerateEmpty is set, a missing result row yields -1 and no error.
func executeID(ctx context.Context, session Executor, query string, params map[string]any, tolerateEmpty bool) (int64, error) {
	id, err := session.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
		result, err := tx.Run(ctx, query, params)
		if err != nil {
			return nil, err
		}
		record, err := optionalRecord(ctx, result)
		if err != nil {
			return nil, err
		}
		if record == nil {
			if tolerateEmpty {
				return int64(-1), nil
			}
			return nil, errors.New("query returned no record")
		}
		id, _, err := neo4j.GetRecordValue[int64](record, "id")
		return id, err
	})
	if err != nil {
		return -1, fmt.Errorf("%w (query: %s)", err, query)
	}
	return id.(int64), nil
}

// quote escapes a Cypher identifier (label, relation type or property key).
func quote(name string) string {
	return "`" + strings.ReplaceAll(name, "`", "``") + "`"
}

// labelString joins labels for use in a pattern; labels may hold ":" separated parts.
func labelString(labels []string) string {
	var parts []string
	for _, label := range labels {
		for _, part := range strings.Split(label, ":") {
			parts = append(parts, quote(part))
		}
	}
	return strings.Join(parts, ":")
}

// Node merges a node by name.
func Node(ctx context.Context, session Executor, labels []string, name string) (int64, error) {
	query := "MERGE (n:" + labelString(labels) + " {name: $name}) RETURN id(n) AS id"
	id, err := executeID(ctx, session, query, map[string]any{"name": name}, false)
	if err != nil {
		return -1, fmt.Errorf("merge node %v: %w", labels, err)
	}
	return id, nil
}

// Query returns the id of a single result row, or -1 if the query has no result row.
func Query(ctx context.Context, session Executor, query string, parameters map[string]any) (int64, error) {
	return executeID(ctx, session, query, parameters, true)
}

// QueryRecord returns the single result record, or nil if there is none.
func QueryRecord(ctx context.Context, session Executor, query string, parameters map[string]any) (*neo4j.Record, error) {
	slog.Debug("Graph QueryRecord", "query", query, "params", parameters)
	record, err := session.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
		result, err := tx.Run(ctx, query, parameters)
		if err != nil {
			return nil, err
		}
		return optionalRecord(ctx, result)
	})
	if err != nil {
		return nil, fmt.Errorf("query record: %w", err)
	}
	return record.(*neo4j.Record), nil
}

// NodeExt merges a node by the match properties and sets additional properties.
func NodeExt(ctx context.Context, session Executor, labels []string, match map[string]string, properties map[string]any) (int64, error) {
	keys := make([]string, 0, len(match))
	for key := range match {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	params := map[string]any{"props": properties}
	matchStrings := make([]string, 0, len(keys))
	for i, key := range keys {
		param := fmt.Sprintf("m%d", i)
		matchStrings = append(matchStrings, quote(key)+": $"+param)
		params[param] = match[key]
	}
	query := "MERGE (n:" + labelString(labels) + " {" + strings.Join(matchStrings, ", ") + "}) " +
		"ON CREATE SET n += $props " +
		"ON MATCH SET n += $props " +
		"RETURN id(n) AS id"
	id, err := executeID(ctx, session, query, params, false)
	if err != nil {
		return -1, fmt.Errorf("merge node %v: %w", labels, err)
	}
	return id, nil
}

const relationMatch = "MATCH (a), (b) WHERE id(a) = $start AND id(b) = $end "

// Relation merges a relation from start to end.
func Relation(ctx context.Context, session Executor, start int64, end int64, labels []string) (int64, error) {
	query := relationMatch +
		"MERGE (a)-[r:" + labelString(labels) + "]->(b) " +
		"RETURN id(r) AS id"
	id, err := executeID(ctx, session, query, map[string]any{"start": start, "end": end}, false)
	if err != nil {
		return -1, fmt.Errorf("merge relation %v: %w", labels, err)
	}
	return id, nil
}

// RelationExt merges a relation from start to end and sets additional properties.
func RelationExt(ctx context.Context, session Executor, start int64, end int64, labels []string, properties map[string]any) (int64, error) {
	query := relationMatch +
		"MERGE (a)-[r:" + labelString(labels) + "]->(b) " +
		"SET r += $props " +
		"RETURN id(r) AS id"
	id, err := executeID(ctx, session, query, map[string]any{"start": start, "end": end, "props": properties}, false)
	if err != nil {
		return -1, fmt.Errorf("merge relation %v: %w", labels, err)
	}
	return id, nil
}

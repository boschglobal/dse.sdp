package graph

import (
	"context"
	"fmt"
	"log/slog"
	"strconv"
	"strings"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

func Driver(dbUri string) (any, error) {
	driver, err := neo4j.NewDriverWithContext(dbUri, neo4j.BasicAuth("", "", ""))
	return driver, err
}

func Close(ctx context.Context) {
	driver := ctx.Value("driver").(neo4j.DriverWithContext)
	if driver != nil {
		driver.Close(ctx)
	}
}

func Session(ctx context.Context) (neo4j.SessionWithContext, error) {
	driver := ctx.Value("driver").(neo4j.DriverWithContext)
	session := driver.NewSession(ctx, neo4j.SessionConfig{
		AccessMode: neo4j.AccessModeWrite,
		//		BoltLogger: neo4j.ConsoleBoltLogger(),  TODO Future CLI option.
	})
	return session, nil
}

func Drop(ctx context.Context, option string) {
    var query string
    // Determine the query based on the option
    switch option {
    case "ast":
        query = "MATCH (n:Ast) DETACH DELETE n"
        fmt.Println("Graph query: MATCH DETACH DELETE Ast nodes")
    case "sim":
        query = "MATCH (n:Sim) DETACH DELETE n"
        fmt.Println("Graph query: MATCH DETACH DELETE Sim nodes")
    case "--all":
        query = "MATCH (n) DETACH DELETE n"
        fmt.Println("Graph query: MATCH DETACH DELETE all nodes")
    default:
        fmt.Println("Incorrect Usage. Use 'ast', 'sim', or '--all'")
        return
    }
    session, _ := Session(ctx)
    defer session.Close(ctx)
    session.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
        return tx.Run(ctx, query, nil)
    })
}

func Node(ctx context.Context, session neo4j.SessionWithContext, labels []string, name string) (int64, error) {
	id, err := session.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
		var b strings.Builder
		b.WriteString("MERGE (n:" + strings.Join(labels, ":") + " {name: $name}) ")
		b.WriteString("RETURN id(n) AS id")
		query := b.String()
		result, err := tx.Run(ctx, query, map[string]any{"name": name})
		if err != nil {
			fmt.Println(query)
			return -1, err
		}
		record, err := result.Single(ctx)
		if err != nil {
			fmt.Println(query)
			return -1, err
		}
		id, _, err := neo4j.GetRecordValue[int64](record, "id")
		if err != nil {
			fmt.Println(query)
		}
		return id, err
	})
	if err != nil {
		fmt.Println("ERROR: adding node:", err)
		return -1, err
	}
	return id.(int64), nil
}

func Query(ctx context.Context, session neo4j.SessionWithContext, query string, parameters map[string]any) (int64, error) {
	// Returns single id ... or -1, error consumed.
	id, err := session.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
		result, err := tx.Run(ctx, query, parameters)
		if err != nil {
			fmt.Println(query)
			return -1, err
		}
		record, err := result.Single(ctx)
		if err != nil {
			return int64(-1), nil // No result, consume the error.
		}
		id, _, err := neo4j.GetRecordValue[int64](record, "id")
		if err != nil {
			fmt.Println(query)
		}
		return id, err
	})
	if err != nil {
		fmt.Println("ERROR: adding node:", err)
		return -1, err
	}
	return id.(int64), nil
}

func QueryRecord(ctx context.Context, session neo4j.SessionWithContext, query string, parameters map[string]any) (*neo4j.Record, error) {
	slog.Info("Graph QueryRecord", "query", query, "params", parameters)
	record, err := session.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
		result, err := tx.Run(ctx, query, parameters)
		if err != nil {
			return nil, err
		}
		record, err := result.Single(ctx)
		if err != nil {
			return nil, nil
		}
		return record, err
	})
	if err != nil {
		slog.Error("Graph QueryRecord", "err", err)
	}
	return record.(*neo4j.Record), err
}

func NodeExt(ctx context.Context, session neo4j.SessionWithContext, labels []string, match map[string]string, properties map[string]any) (int64, error) {
	// Create a Node with additional properties.
	id, err := session.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
		matchStrings := []string{}
		for name, value := range match {
			matchStrings = append(matchStrings, fmt.Sprintf("%s: '%s'", name, value))
		}
		var b strings.Builder
		b.WriteString("MERGE (n:" + strings.Join(labels, ":") + "{ ")
		b.WriteString(strings.Join(matchStrings, ","))
		b.WriteString("}) ")
		b.WriteString("ON CREATE SET n += $props ")
		b.WriteString("ON MATCH SET n += $props ")
		b.WriteString("RETURN id(n) AS id")
		query := b.String()
		result, err := tx.Run(ctx, query, map[string]any{"match": match, "props": properties})
		if err != nil {
			fmt.Println(query)
			return -1, err
		}
		record, err := result.Single(ctx)
		if err != nil {
			fmt.Println(query)
			return -1, err
		}
		id, _, err := neo4j.GetRecordValue[int64](record, "id")
		if err != nil {
			fmt.Println(query)
		}
		return id, err
	})
	if err != nil {
		fmt.Println("ERROR: adding node:", err)
		return -1, err
	}
	return id.(int64), nil
}

func Relation(ctx context.Context, session neo4j.SessionWithContext, start int64, end int64, labels []string) (int64, error) {
	id, err := session.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
		var b strings.Builder
		b.WriteString("MATCH (a), (b) ")
		b.WriteString("WHERE ")
		b.WriteString("    id(a) = " + strconv.FormatInt(start, 10) + " ")
		b.WriteString("AND ")
		b.WriteString("    id(b) = " + strconv.FormatInt(end, 10) + " ")
		b.WriteString("MERGE (a)-[r:" + strings.Join(labels, ":") + "]->(b) ")
		b.WriteString("RETURN id(r) AS id")
		query := b.String()
		result, err := tx.Run(ctx, query, map[string]any{})
		if err != nil {
			fmt.Println(query)
			return -1, err
		}
		record, err := result.Single(ctx)
		if err != nil {
			fmt.Println(query)
			return -1, err
		}
		id, _, err := neo4j.GetRecordValue[int64](record, "id")
		if err != nil {
			fmt.Println(query)
		}
		return id, err
	})
	if err != nil {
		fmt.Println("ERROR: adding relation:", err)
		return -1, err
	}
	return id.(int64), nil
}

func RelationExt(ctx context.Context, session neo4j.SessionWithContext, start int64, end int64, labels []string, properties map[string]any) (int64, error) {
	id, err := session.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
		var b strings.Builder
		b.WriteString("MATCH (a), (b) ")
		b.WriteString("WHERE ")
		b.WriteString("    id(a) = " + strconv.FormatInt(start, 10) + " ")
		b.WriteString("AND ")
		b.WriteString("    id(b) = " + strconv.FormatInt(end, 10) + " ")
		b.WriteString("MERGE (a)-[r:" + strings.Join(labels, ":") + "]->(b) ")
		b.WriteString("SET r += $props ")
		b.WriteString("RETURN id(r) AS id")
		query := b.String()
		result, err := tx.Run(ctx, query, map[string]any{"props": properties})
		if err != nil {
			fmt.Println(query)
			return -1, err
		}
		record, err := result.Single(ctx)
		if err != nil {
			fmt.Println(query)
			return -1, err
		}
		id, _, err := neo4j.GetRecordValue[int64](record, "id")
		if err != nil {
			fmt.Println(query)
		}
		return id, err
	})
	if err != nil {
		fmt.Println("ERROR: adding relation:", err)
		return -1, err
	}
	return id.(int64), nil
}

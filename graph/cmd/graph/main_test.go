//go:build test_e2e
// +build test_e2e

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
	"testing"

	"github.com/boschglobal/dse.sdp/graph/internal/pkg/graph"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
	"github.com/rogpeppe/go-internal/testscript"
)

func TestMain(m *testing.M) {
	os.Exit(testscript.RunMain(m, map[string]func() int{
		"graph": main_,
	}))
}

func TestE2E(t *testing.T) {
	testscript.Run(t, testscript.Params{
		Dir: "testdata/reports",
		Cmds: map[string]func(ts *testscript.TestScript, neg bool, args []string){
			"fubar": func(ts *testscript.TestScript, neg bool, args []string) {
				fmt.Fprint(ts.Stdout(), "hello world")
			},
			"filecontains": func(ts *testscript.TestScript, neg bool, args []string) {
				if len(args) != 2 {
					ts.Fatalf("filecontains <file> <text>")
				}
				got := ts.ReadFile(args[0])
				want := args[1]
				if strings.Contains(got, want) == neg {
					ts.Fatalf("filecontains %q; %q not found in file:\n%q", args[0], want, got)
				}
			},
			"graphquery": func(ts *testscript.TestScript, neg bool, args []string) {
				db := "bolt://localhost:7687"
				query := ""
				if len(args) == 1 {
					query = args[0]
				} else if len(args) == 2 {
					db = args[0]
					query = args[1]
				} else {
					ts.Fatalf("graph <query> OR graph <db> <query>")
				}
				driver, err := graph.Driver(db)
				if err != nil {
					ts.Fatalf("Graph driver error: %+v", err)
				}
				ctx := context.WithValue(context.Background(), "driver", driver)
				defer graph.Close(ctx)
				session, err := graph.Session(ctx)
				if err != nil {
					ts.Fatalf("Graph session error: %+v", err)
				}
				defer session.Close(ctx)
				session.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
					return tx.Run(ctx, query, map[string]any{})
				})

			},
			"graphq": func(ts *testscript.TestScript, neg bool, args []string) {
				if len(args) < 2 || len(args) > 3 {
					ts.Fatalf("Usage: graphq <optional-count> <query-file> <json-params>")
				}

				var queryFile, jsonParams string
				var expectedCount int
				
				if len(args) == 2 {
					queryFile = args[0]
					jsonParams = args[1]
				} else if len(args) == 3 {
					queryFile = args[0]
					jsonParams = args[2]
					var err error
					expectedCount, err = strconv.Atoi(args[1])
					if err != nil {
						ts.Fatalf("Failed to read count value: %v", err)
					}
				}

				// Read the query from the .cyp file.
				query, err := os.ReadFile(queryFile)
				if err != nil {
					ts.Fatalf("Failed to read query file %s: %v", queryFile, err)
				}

				// Parse the JSON string into a map.
				var params map[string]interface{}
				err = json.Unmarshal([]byte(jsonParams), &params)
				if err != nil {
					ts.Fatalf("Failed to parse JSON params: %v", err)
				}

				// Connect to the database.
				driver, err := graph.Driver("bolt://localhost:7687")
				if err != nil {
					ts.Fatalf("Graph driver error: %+v", err)
				}
				ctx := context.WithValue(context.Background(), "driver", driver)
				defer graph.Close(ctx)

				session, err := graph.Session(ctx)
				if err != nil {
					ts.Fatalf("Graph session error: %+v", err)
				}
				defer session.Close(ctx)

				// Execute the query with parameters.
				records, err := session.ExecuteRead(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
					result, err := tx.Run(ctx, string(query), params)
					if err != nil {
						return nil, err
					}
					return result.Collect(ctx)
				})
				if err != nil {
					ts.Fatalf("Failed to execute query: %+v", err)
				}

				// Count the number of matches.
				matchCount := len(records.([]*neo4j.Record))
				fmt.Printf("Graph query '%s' returned %d matches with params %+v.\n", query, matchCount, params)

				if len(args) == 3 && matchCount != expectedCount {
					ts.Fatalf("Expected %d match but got %d", expectedCount, matchCount)
				}
			},
		},
	})
}

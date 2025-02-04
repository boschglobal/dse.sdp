//go:build test_e2e
// +build test_e2e

package kind

import (
	"context"
	"fmt"
	"os"
	"testing"
	"encoding/json"
	"strconv"
	"path/filepath"

	"github.com/boschglobal/dse.sdp/graph/internal/pkg/graph"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
	"github.com/rogpeppe/go-internal/testscript"
)

func TestMain(m *testing.M) {
	os.Exit(m.Run())
}

func TestKindE2E(t *testing.T) {
    dirs := []string{"testdata"}
    
    for _, dir := range dirs {
        t.Run(filepath.Base(dir), func(t *testing.T) {
            testscript.Run(t, testscript.Params{
                Dir: dir,
                Cmds: map[string]func(ts *testscript.TestScript, neg bool, args []string){
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

                    "graphdrop": func(ts *testscript.TestScript, neg bool, args []string) {
                        option := args[0]
                        if option != "--all" && option != "ast" && option != "sim" {
                            ts.Fatalf("Invalid option: %s. Valid options are '--all', 'ast', or 'sim'", option)
                        }
                        ctx := context.Background()

                        // Connect to the database.
                        driver, err := graph.Driver("bolt://localhost:7687")
                        if err != nil {
                            ts.Fatalf("Failed to connect to graph database: %+v", err)
                        }
                        ctx = context.WithValue(ctx, "driver", driver)
                        defer graph.Close(ctx)

                        // Call the DropAll function.
                        graph.Drop(ctx, option)
                        fmt.Fprint(ts.Stdout(), "All nodes dropped successfully")
                    },

                    "import": func(ts *testscript.TestScript, neg bool, args []string) {
                        if len(args) != 1 {
                            ts.Fatalf("Usage: import <yaml-file>")
                        }
                        file := args[0]

                        handler := &YamlKindHandler{}
                        data := handler.Detect(file)
                        if data == nil {
                            ts.Fatalf("Failed to detect YAML data from file: %s", file)
                        }

                        ctx := context.Background()

                        // Connect to the database.
                        driver, err := graph.Driver("bolt://localhost:7687")
                        if err != nil {
                            ts.Fatalf("Failed to connect to graph database: %+v", err)
                        }
                        ctx = context.WithValue(ctx, "driver", driver)
                        defer graph.Close(ctx)

                        // Call the Import function.
                        handler.Import(ctx, file, data)
                        fmt.Fprint(ts.Stdout(), "Import successful")
                    },

                    "report": func(ts *testscript.TestScript, neg bool, args []string) {
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

                        // Query to count each node type.
                        query := `
                            MATCH (n)
                            RETURN DISTINCT labels(n) AS labels, count(n) AS count
                        `

                        // Execute the count query.
                        records, err := session.ExecuteRead(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
                            result, err := tx.Run(ctx, query, nil)
                            if err != nil {
                                return nil, err
                            }
                            return result.Collect(ctx)
                        })
                        if err != nil {
                            ts.Fatalf("Failed to execute count query: %+v", err)
                        }

                        // Print the counts for each node type.
                        for _, record := range records.([]*neo4j.Record) {
                            labelsValue, ok := record.Get("labels")
                            if !ok {
                                ts.Fatalf("Missing labels in record")
                            }
                            labels := labelsValue.([]interface{})

                            countValue, ok := record.Get("count")
                            if !ok {
                                ts.Fatalf("Missing count in record")
                            }
                            count := countValue.(int64)

                            fmt.Printf("Node type: %v, Count: %d\n", labels, count)
                        }
                    },
                },
            })
        })
    }
}

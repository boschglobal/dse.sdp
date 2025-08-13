package graph

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"

	"github.com/boschglobal/dse.clib/extra/go/command"
	"github.com/boschglobal/dse.clib/extra/go/command/log"

	"github.com/boschglobal/dse.sdp/graph/internal/pkg/file/kind"
	"github.com/boschglobal/dse.sdp/graph/internal/pkg/graph"
)

type GraphImportCommand struct {
	command.Command
	logLevel  int
	optImport string
	optDb     string
}

func NewGraphImportCommand(name string) *GraphImportCommand {
	c := &GraphImportCommand{
		Command: command.Command{
			Name:    name,
			FlagSet: flag.NewFlagSet(name, flag.ExitOnError),
		},
	}
	c.FlagSet().IntVar(&c.logLevel, "log", 4, "Loglevel")
	c.FlagSet().StringVar(&c.optImport, "import", "", "import files to the database")
	c.FlagSet().StringVar(&c.optDb, "db", "bolt://localhost:7687", "database connection string")
	return c
}

// CommandRunner interface functions.
func (c GraphImportCommand) Name() string {
	return c.Command.Name
}

func (c GraphImportCommand) FlagSet() *flag.FlagSet {
	return c.Command.FlagSet
}

func (c *GraphImportCommand) Parse(args []string) error {
	return c.FlagSet().Parse(args)
}

func (c *GraphImportCommand) Run() error {
	slog.SetDefault(log.NewLogger(c.logLevel))
	slog.Info("Connect to graph", "db", c.optDb)
	ctx := context.Background()
	driver, err := graph.Driver(c.optDb)
	if err != nil {
		slog.Info("Graph driver error", "error", err)
		return err
	}
	ctx = context.WithValue(ctx, "driver", driver)
	defer graph.Close(ctx)

	session, err := graph.Session(ctx)
	if err != nil {
		slog.Info("Graph session error", "err", err)
		return err
	}
	defer session.Close(ctx)

	args := c.FlagSet().Args() // Get positional arguments
	if len(args) == 0 {
		slog.Error("Usage: graph import <yaml-file>")
		return nil
	}
	file := args[0]
	c.importFiles(ctx, file, session)

	return nil
}

func (c *GraphImportCommand) matchNode(ctx context.Context, session neo4j.SessionWithContext) {
	match_instance := `
	MATCH (ast_mi:Ast:ModelInst), (sim_mi:Sim:ModelInst)
    WHERE ast_mi.model_name = sim_mi.name
    MERGE (sim_mi)-[:Represents]->(ast_mi)
    `
	_, err := graph.Query(ctx, session, match_instance, nil)
	if err != nil {
		slog.Info("Failed to create relationship", "error", err)
	}

	match_channel := `
	MATCH (ast_ch:Ast:SimulationChannel), (sim_ch:Sim:Channel)
	WHERE ast_ch.channel_name = sim_ch.name
	MERGE (sim_ch)-[:Represents]->(ast_ch)
	`
	_, err = graph.Query(ctx, session, match_channel, nil)
	if err != nil {
		slog.Info("Failed to create relationship", "error", err)
	}
}

func (c *GraphImportCommand) importFiles(ctx context.Context, path string, session neo4j.SessionWithContext) {
	if path == "" {
		slog.Error("Usage: import <yaml-path-or-file>")
		return
	}

	// Detect all YAML files in the path recursively.
	var yamlFiles []string
	err := filepath.Walk(path, func(p string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && (strings.HasSuffix(p, ".yaml") || strings.HasSuffix(p, ".yml")) {
			yamlFiles = append(yamlFiles, p)
		}
		return nil
	})
	if err != nil {
		slog.Error("Error walking the path", "error", err)
		return
	}

	if len(yamlFiles) == 0 {
		slog.Error("No YAML files found in path", "path", path)
		return
	}

	// Initialize the YAML handler and graph driver.
	handler := &kind.YamlKindHandler{}
	driver, err := graph.Driver(c.optDb)
	if err != nil {
		slog.Info("Failed to connect to graph database", "error", err)
		return
	}
	ctx = context.WithValue(ctx, "driver", driver)
	defer graph.Close(ctx)

	// Import each YAML file
	for _, yamlFile := range yamlFiles {
		slog.Info("Importing file", "file", yamlFile)
		data := handler.Detect(yamlFile)
		handler.Import(ctx, yamlFile, data)
	}

	// Create additional relations after all files are imported
	c.createRelationships(ctx, session)

}

func (c *GraphImportCommand) createRelationships(ctx context.Context, session neo4j.SessionWithContext) {
	query_InstanceOf := `
	MATCH (inst:ModelInst), (m:Model)
	WHERE inst.model = m.name
	MERGE (inst)-[:InstanceOf]->(m)`
	_, _ = graph.Query(ctx, session, query_InstanceOf, nil)

	query_Belongs := `
	MATCH (sc:SimbusChannel)
	MATCH (c:Channel)
	WHERE sc.name = c.name
	MERGE (c)-[:Belongs]->(sc)`
	_, _ = graph.Query(ctx, session, query_Belongs, nil)

	query_Selects := `
	MATCH (sg:SignalGroup)-[sgHas:Has]->(l:Label)
	MATCH (mi:ModelInst)-[miHas:Has]->(sl:Selector)
	WHERE sl.selectorName = l.label_name AND sl.selectorValue = l.label_value
	WITH sg, sl, l, mi, COUNT(miHas) AS miCount, COUNT(sgHas) AS sgCount
	WHERE miCount = sgCount
	MERGE (sl)-[:Selects]->(l)`
	_, _ = graph.Query(ctx, session, query_Selects, nil)

	query_SelectorCount := `
	CALL {
		MATCH (c:Channel)<-[id:Identifies]-(s:Selector)<-[has:Has]-(m:Model)<-[i:InstanceOf]-(mi:ModelInst{name:$mi_name})
		RETURN s, c, id
		UNION
		MATCH (m:Model)<-[i:InstanceOf]-(mi:ModelInst{name:$mi_name})-[h:Has]->(s:Selector)-[id:Identifies]->(c:Channel)
		RETURN s, c, id
	}
	WITH c as channel, s as selector
	RETURN channel, count(selector) AS selectorCount
	`

	query_LabelCount := `
	CALL {
		MATCH (c:Channel)<-[id:Identifies]-(s:Selector)<-[has:Has]-(m:Model)<-[i:InstanceOf]-(mi:ModelInst{name:$mi_name})
		RETURN s, c, id
		UNION
		MATCH (m:Model)<-[i:InstanceOf]-(mi:ModelInst{name:$mi_name})-[h:Has]->(s:Selector)-[id:Identifies]->(c:Channel)
		RETURN s, c, id
	}
	WITH c as channel, s as selector
	MATCH (channel)<-[:Identifies]-(selector)-[selects:Selects]->(l:Label)<-[h:Has]-(sig:SignalGroup)
	RETURN channel, sig, count(l) AS labelCount
	`

	// Get all model instance names.
	result, err := session.Run(ctx, `MATCH (mi:ModelInst) RETURN mi.name AS mi_name`, nil)
	if err != nil {
		slog.Error(fmt.Sprintf("Failed to get model instance names: %v", err))
	}

	for result.Next(ctx) {
		miNameVal, _ := result.Record().Get("mi_name")
		miName, ok := miNameVal.(string)
		if !ok {
			continue
		}

		mi_properties := map[string]any{"mi_name": miName}

		selectorCounts := make(map[int64]int64)
		channelNodeIDs := make(map[int64]int64)

		result1, _ := session.Run(ctx, query_SelectorCount, mi_properties)
		for result1.Next(ctx) {
			record := result1.Record()
			if selectorCount, _ := record.Get("selectorCount"); selectorCount != nil {
				if count, ok := selectorCount.(int64); ok {
					if channelValue, exists := record.Get("channel"); exists {
						if channelNode, ok := channelValue.(neo4j.Node); ok {
							channelID := channelNode.Id
							selectorCounts[channelID] = count
							channelNodeIDs[channelID] = channelID
						}
					}
				}
			}
		}

		result2, _ := session.Run(ctx, query_LabelCount, mi_properties)
		for result2.Next(ctx) {
			record := result2.Record()
			labelCount, _ := record.Get("labelCount")
			sgValue, _ := record.Get("sig")
			channelValue, _ := record.Get("channel")

			if labelCount, ok := labelCount.(int64); ok {
				if sgNode, ok := sgValue.(neo4j.Node); ok {
					if chNode, ok := channelValue.(neo4j.Node); ok {
						channelID, signalGroupID := chNode.Id, sgNode.Id
						if selectorCount, found := selectorCounts[channelID]; found {
							if labelCount >= selectorCount {
								relationshipQuery := `
								MATCH (c:Channel) WHERE ID(c) = $channelID
								MATCH (sg:SignalGroup) WHERE ID(sg) = $signalGroupID
								MERGE (c)-[:Represents]->(sg)`
								params := map[string]any{
									"channelID":     channelID,
									"signalGroupID": signalGroupID,
								}
								_, err := session.Run(ctx, relationshipQuery, params)
								if err != nil {
									slog.Info(fmt.Sprintf("Failed to create relationship: %v", err))
								} else {
									slog.Info(fmt.Sprintf("Relationship created: Channel(ID: %d) -> Represents -> SignalGroup(ID: %d)", channelID, signalGroupID))
								}
							} else {
								slog.Info(fmt.Sprintf("Skipping channel ID %d (labelCount: %d, selectorCount: %d)", channelID, labelCount, selectorCount))
							}
						}
					}
				}
			}
		}
	}
}

package graph

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"

	"github.com/boschglobal/dse.sdp/graph/internal/pkg/file/kind"
	"github.com/boschglobal/dse.sdp/graph/internal/pkg/graph"
)

const (
	queryInstanceOf = `
	MATCH (inst:ModelInst), (m:Model)
	WHERE inst.model = m.name
	MERGE (inst)-[:InstanceOf]->(m)`

	queryBelongs = `
	MATCH (sc:SimbusChannel)
	MATCH (c:Channel)
	WHERE sc.name = c.name
	MERGE (c)-[:Belongs]->(sc)`

	querySelects = `
	MATCH (sg:SignalGroup)-[sgHas:Has]->(l:Label)
	MATCH (mi:ModelInst)-[miHas:Has]->(sl:Selector)
	WHERE sl.selectorName = l.label_name AND sl.selectorValue = l.label_value
	WITH sg, sl, l, mi, COUNT(miHas) AS miCount, COUNT(sgHas) AS sgCount
	WHERE miCount = sgCount
	MERGE (sl)-[:Selects]->(l)`

	queryModelInstNames = `MATCH (mi:ModelInst) RETURN mi.name AS mi_name`

	querySelectorCount = `
	CALL {
		MATCH (c:Channel)<-[id:Identifies]-(s:Selector)<-[has:Has]-(m:Model)<-[i:InstanceOf]-(mi:ModelInst{name:$mi_name})
		RETURN s, c, id
		UNION
		MATCH (m:Model)<-[i:InstanceOf]-(mi:ModelInst{name:$mi_name})-[h:Has]->(s:Selector)-[id:Identifies]->(c:Channel)
		RETURN s, c, id
	}
	WITH c as channel, s as selector
	RETURN id(channel) AS channel_id, count(selector) AS selectorCount
	`

	queryLabelCount = `
	CALL {
		MATCH (c:Channel)<-[id:Identifies]-(s:Selector)<-[has:Has]-(m:Model)<-[i:InstanceOf]-(mi:ModelInst{name:$mi_name})
		RETURN s, c, id
		UNION
		MATCH (m:Model)<-[i:InstanceOf]-(mi:ModelInst{name:$mi_name})-[h:Has]->(s:Selector)-[id:Identifies]->(c:Channel)
		RETURN s, c, id
	}
	WITH c as channel, s as selector
	MATCH (channel)<-[:Identifies]-(selector)-[selects:Selects]->(l:Label)<-[h:Has]-(sig:SignalGroup)
	RETURN id(channel) AS channel_id, id(sig) AS sig_id, count(l) AS labelCount
	`

	queryRepresents = `
	MATCH (c:Channel) WHERE ID(c) = $channelID
	MATCH (sg:SignalGroup) WHERE ID(sg) = $signalGroupID
	MERGE (c)-[:Represents]->(sg)`
)

type GraphImportCommand struct {
	base
}

func NewGraphImportCommand(name string) *GraphImportCommand {
	c := &GraphImportCommand{}
	c.init(name)
	return c
}

func (c *GraphImportCommand) Run() error {
	c.setupLogging()

	args := c.FlagSet().Args()
	if len(args) == 0 {
		return errors.New("usage: graph import <yaml-path-or-file>")
	}

	ctx, session, closeFn, err := c.connect()
	if err != nil {
		return err
	}
	defer closeFn()

	return importFiles(ctx, args[0], session)
}

func findYAMLFiles(path string) ([]string, error) {
	var files []string
	err := filepath.WalkDir(path, func(p string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() && isYAMLFile(p) {
			files = append(files, p)
		}
		return nil
	})
	return files, err
}

func isYAMLFile(name string) bool {
	return strings.HasSuffix(name, ".yaml") || strings.HasSuffix(name, ".yml")
}

func importKindFiles(ctx context.Context, handler *kind.YamlKindHandler, files []string) error {
	var errs []error
	for _, file := range files {
		slog.Info("Importing file", "file", file)
		data := handler.Detect(file)
		if data == nil {
			slog.Warn("No kind documents found", "file", file)
			continue
		}
		if err := handler.Import(ctx, file, data); err != nil {
			errs = append(errs, err)
		}
	}
	return errors.Join(errs...)
}

func importFiles(ctx context.Context, path string, session neo4j.SessionWithContext) error {
	yamlFiles, err := findYAMLFiles(path)
	if err != nil {
		return fmt.Errorf("walk path %q: %w", path, err)
	}
	if len(yamlFiles) == 0 {
		return fmt.Errorf("no YAML files found in path %q", path)
	}

	importErr := importKindFiles(ctx, &kind.YamlKindHandler{}, yamlFiles)

	// Create additional relations after all files are imported.
	createRelationships(ctx, session)
	return importErr
}

func createRelationships(ctx context.Context, session neo4j.SessionWithContext) {
	for _, query := range []string{queryInstanceOf, queryBelongs, querySelects} {
		if _, err := graph.Query(ctx, session, query, nil); err != nil {
			slog.Error("Failed to create relationships", "error", err)
		}
	}

	names, err := modelInstanceNames(ctx, session)
	if err != nil {
		slog.Error("Failed to get model instance names", "error", err)
		return
	}
	for _, name := range names {
		if err := representChannels(ctx, session, name); err != nil {
			slog.Error("Failed to create Represents relationships", "model_inst", name, "error", err)
		}
	}
}

func modelInstanceNames(ctx context.Context, session neo4j.SessionWithContext) ([]string, error) {
	result, err := session.Run(ctx, queryModelInstNames, nil)
	if err != nil {
		return nil, err
	}
	var names []string
	for result.Next(ctx) {
		if name, ok := recordValue[string](result.Record(), "mi_name"); ok {
			names = append(names, name)
		}
	}
	return names, result.Err()
}

func representChannels(ctx context.Context, session neo4j.SessionWithContext, miName string) error {
	params := map[string]any{"mi_name": miName}

	selectorCounts, err := selectorCountsByChannel(ctx, session, params)
	if err != nil {
		return err
	}

	result, err := session.Run(ctx, queryLabelCount, params)
	if err != nil {
		return err
	}
	for result.Next(ctx) {
		record := result.Record()
		labelCount, ok1 := recordValue[int64](record, "labelCount")
		signalGroupID, ok2 := recordValue[int64](record, "sig_id")
		channelID, ok3 := recordValue[int64](record, "channel_id")
		if !ok1 || !ok2 || !ok3 {
			continue
		}
		selectorCount, found := selectorCounts[channelID]
		if !found {
			continue
		}
		if labelCount < selectorCount {
			slog.Info(fmt.Sprintf("Skipping channel ID %d (labelCount: %d, selectorCount: %d)", channelID, labelCount, selectorCount))
			continue
		}
		_, err := session.Run(ctx, queryRepresents, map[string]any{
			"channelID":     channelID,
			"signalGroupID": signalGroupID,
		})
		if err != nil {
			slog.Info(fmt.Sprintf("Failed to create relationship: %v", err))
			continue
		}
		slog.Info(fmt.Sprintf("Relationship created: Channel(ID: %d) -> Represents -> SignalGroup(ID: %d)", channelID, signalGroupID))
	}
	return result.Err()
}

func selectorCountsByChannel(ctx context.Context, session neo4j.SessionWithContext, params map[string]any) (map[int64]int64, error) {
	result, err := session.Run(ctx, querySelectorCount, params)
	if err != nil {
		return nil, err
	}
	counts := make(map[int64]int64)
	for result.Next(ctx) {
		record := result.Record()
		count, ok1 := recordValue[int64](record, "selectorCount")
		channelID, ok2 := recordValue[int64](record, "channel_id")
		if ok1 && ok2 {
			counts[channelID] = count
		}
	}
	return counts, result.Err()
}

func recordValue[T any](record *neo4j.Record, key string) (T, bool) {
	v, ok := record.Get(key)
	if !ok {
		var zero T
		return zero, false
	}
	t, ok := v.(T)
	return t, ok
}

package kind

import (
	"context"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"

	"github.com/boschglobal/dse.sdp/graph/internal/pkg/graph"
	"github.com/boschglobal/dse.schemas/code/go/dse/kind"
)

type SignalGroupSpec kind.SignalGroupSpec

func newSignalGroupSpec() *SignalGroupSpec {
	return new(SignalGroupSpec)
}

func (sg *SignalGroupSpec) MergeGraph(ctx context.Context, session neo4j.SessionWithContext, kd *KindDoc) error {
	signalGroupID := kd.kind_id

	// SIGNALGROUP -[HAS]-> SIGNAL
	for _, s := range sg.Signals {
		matchProps := map[string]string{
			"name": s.Signal,
		}
		nodeProps := map[string]any{
			"annotations": s.Annotations,
		}
		signalID, _ := graph.NodeExt(ctx, session, []string{"Sim:Signal"}, matchProps, nodeProps)
		graph.Relation(ctx, session, signalGroupID, signalID, []string{"Contains"})
	}

	// SIGNALGROUP -[HAS]-> LABEL
	for name, value := range kd.Metadata.Labels {
		properties := map[string]any{
			"signalgroup_id": signalGroupID,
			"label_name":     name,
			"label_value":    value,
			"props": map[string]any{
				"label_value": value,
			},
		}

		query := `
			MATCH (sg) WHERE id(sg) = $signalgroup_id
			MERGE (sg)-[:Represents]->(l:Sim:Label {label_name: $label_name})
			ON CREATE SET l += $props
			ON MATCH SET l += $props
			RETURN id(l) AS id
		`
		_, _ = graph.Query(ctx, session, query, properties)

		Query_Selects := `
		MATCH (mi:ModelInst)-[has:Has]->(s:Selector)
		MATCH (ModelInst)-[miHas:Has]->(s:Selector {selectorName: $label_name, selectorValue: $label_value}),
			(SignalGroup)-[sgHas:Represents]->(l:Label {label_name: $label_name, label_value: $label_value})
		WITH s, l, count(miHas) AS miCount, count(sgHas) AS sgCount
		WHERE miCount = sgCount
		MERGE (s)-[:Selects]->(l)
		`
		_, _ = graph.Query(ctx, session, Query_Selects, properties)

		query_HasSelector := `
		// Match Model and ModelInst if they exist
		OPTIONAL MATCH (m:Model)
		OPTIONAL MATCH (mi:ModelInst)

		// Case 1: If Model exists
		FOREACH (_ IN CASE WHEN m IS NOT NULL AND mi IS NULL THEN [1] ELSE [] END |
			CREATE (sel:Selector {selectorName: $label_name, selectorValue: $label_value})
			MERGE (m)-[:Has]->(sel)
			FOREACH (_ IN CASE WHEN mi.model = m.name THEN [1] ELSE [] END |
				MERGE (mi)-[:Has]->(sel)
			)
		)

		// Case 2: If ModelInst exists
		FOREACH (_ IN CASE WHEN mi IS NOT NULL AND m IS NULL THEN [1] ELSE [] END |
			MERGE (sel:Selector {selectorName: $label_name, selectorValue: $label_value})
			MERGE (mi)-[:Has]->(sel)
				FOREACH (_ IN CASE WHEN mi.model = m.name THEN [1] ELSE [] END |
				MERGE (m)-[:Has]->(sel)
			)
		)
		`
		// Execute the query with the provided properties
		_, _ = graph.Query(ctx, session, query_HasSelector, properties)
	}

	return nil
}

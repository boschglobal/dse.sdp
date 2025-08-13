package kind

import (
	"context"
	"strconv"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"

	"github.com/boschglobal/dse.schemas/code/go/dse/kind"
	"github.com/boschglobal/dse.sdp/graph/internal/pkg/graph"
)

type SignalGroupSpec kind.SignalGroupSpec

func newSignalGroupSpec() *SignalGroupSpec {
	return new(SignalGroupSpec)
}

func (sg *SignalGroupSpec) MergeGraph(ctx context.Context, session neo4j.SessionWithContext, kd *KindDoc) error {
	signalGroupID := kd.kind_id

	// Convert signalGroupID to string
	signalGroupIDStr := strconv.FormatInt(signalGroupID, 10)

	// SIGNALGROUP -[CONTAINS]-> SIGNAL
	for _, s := range sg.Signals {
		matchProps := map[string]string{
			"name":            s.Signal,
			"signal_group_id": signalGroupIDStr, // Convert int64 to string
			// "annotations": *s.Annotations,
		}
		nodeProps := map[string]any{}
		if s.Annotations != nil {
			nodeProps["annotations"] = *s.Annotations
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
			MERGE (sg)-[:Has]->(l:Sim:Label {label_name: $label_name})
			ON CREATE SET l += $props
			ON MATCH SET l += $props
			RETURN id(l) AS id
		`
		_, _ = graph.Query(ctx, session, query, properties)
	}

	// Create nodes based on Graph fragment
	if graphAnnotations, ok := kd.Metadata.Annotations["graph"].(map[string]interface{}); ok {
		var edgeLabel, edgeDirection string
		if edge, ok := graphAnnotations["edge"].(map[string]interface{}); ok {
			if lbl, ok := edge["label"].(string); ok {
				edgeLabel = lbl
			}
			if dir, ok := edge["direction"].(string); ok {
				edgeDirection = dir
			}
		}

		if nodes, ok := graphAnnotations["nodes"].([]interface{}); ok {
			for _, node := range nodes {
				if nodeMap, ok := node.(map[string]interface{}); ok {
					labels := []string{}
					properties := map[string]interface{}{}

					if lbl, ok := nodeMap["label"].(string); ok {
						labels = append(labels, "Sim:"+lbl)
					}
					if props, ok := nodeMap["properties"].(map[string]interface{}); ok {
						properties = props
					}

					matchProps := map[string]string{}
					nodeProps := map[string]any{}

					for k, v := range properties {
						switch v := v.(type) {
						case string:
							matchProps[k] = v
						default:
							nodeProps[k] = v
						}
					}

					newNodeID, _ := graph.NodeExt(ctx, session, labels, matchProps, nodeProps)
					if edgeDirection == "in" {
						signalGroupID, newNodeID = newNodeID, signalGroupID
					}
					graph.Relation(ctx, session, signalGroupID, newNodeID, []string{edgeLabel})
				}
			}
		}
	}

	return nil
}

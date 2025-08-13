package kind

import (
	"context"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"

	"github.com/boschglobal/dse.schemas/code/go/dse/kind"
	"github.com/boschglobal/dse.sdp/graph/internal/pkg/graph"
)

type ModelSpec kind.ModelSpec

func newModelSpec() *ModelSpec {
	return new(ModelSpec)
}

func (m *ModelSpec) MergeGraph(ctx context.Context, session neo4j.SessionWithContext, kd *KindDoc) error {
	model_id := kd.kind_id

	if m.Channels != nil {
		for _, c := range *m.Channels {
			// Handle Selectors
			if c.Selectors != nil && c.Alias != nil {
				for key, value := range *c.Selectors {
					selector_match_props := map[string]string{
						"channelAlias":  *c.Alias,
						"selectorName":  key,
						"selectorValue": value,
					}
					selector_node_props := map[string]any{}
					selector_id, _ := graph.NodeExt(ctx, session, []string{"Sim:Selector"}, selector_match_props, selector_node_props)

					// Relation between Model and Selector
					graph.Relation(ctx, session, model_id, selector_id, []string{"Has"})

				}
			}
		}
	}

	// MODEL -[HAS]-> RUNTIME
	var runtime_id int64
	if m.Runtime != nil {
		runtime_match_props := map[string]string{}
		runtime_node_props := map[string]any{}
		runtime_id, _ = graph.NodeExt(ctx, session, []string{"Sim:Runtime"}, runtime_match_props, runtime_node_props)
		graph.Relation(ctx, session, model_id, runtime_id, []string{"Has"})
	}

	// RUNTIME -[HAS]-> DYNLIB
	if m.Runtime != nil && m.Runtime.Dynlib != nil {
		for _, d := range *m.Runtime.Dynlib {
			match_props := map[string]string{
				"arch": *d.Arch,
				"os":   *d.Os,
				"path": d.Path,
			}
			node_props := map[string]any{
				"annotations": d.Annotations,
				"libs":        d.Libs,
				"variant":     d.Variant,
			}
			dynlib_id, _ := graph.NodeExt(ctx, session, []string{"Sim:Dynlib"}, match_props, node_props)
			graph.Relation(ctx, session, runtime_id, dynlib_id, []string{"Has"})
		}
	}

	// RUNTIME -[HAS]-> EXECUTABLE
	if m.Runtime != nil && m.Runtime.Executable != nil {
		for _, e := range *m.Runtime.Executable {
			match_props := map[string]string{
				"arch": *e.Arch,
				"os":   *e.Os,
			}
			node_props := map[string]any{
				"annotations": e.Annotations,
				"libs":        e.Libs,
			}
			executable_id, _ := graph.NodeExt(ctx, session, []string{"Sim:Executable"}, match_props, node_props)
			graph.Relation(ctx, session, runtime_id, executable_id, []string{"Has"})
		}
	}

	// RUNTIME -[HAS]-> GATEWAY
	if m.Runtime != nil && m.Runtime.Gateway != nil {
		match_props := map[string]string{}
		node_props := map[string]any{
			"annotations": m.Runtime.Gateway.Annotations,
		}
		gateway_id, _ := graph.NodeExt(ctx, session, []string{"Sim:Gateway"}, match_props, node_props)
		graph.Relation(ctx, session, runtime_id, gateway_id, []string{"Has"})
	}

	// RUNTIME -[HAS]-> MCL
	if m.Runtime != nil && m.Runtime.Mcl != nil {
		for _, mcl := range *m.Runtime.Mcl {
			match_props := map[string]string{
				"arch": *mcl.Arch,
				"os":   *mcl.Os,
				"path": mcl.Path,
			}
			node_props := map[string]any{
				"annotations": mcl.Annotations,
				"libs":        mcl.Libs,
				"variant":     mcl.Variant,
			}
			mcl_id, _ := graph.NodeExt(ctx, session, []string{"Sim:Mcl"}, match_props, node_props)
			graph.Relation(ctx, session, runtime_id, mcl_id, []string{"Has"})
		}
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

					matchProps := make(map[string]string)
					nodeProps := make(map[string]any)

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
						model_id, newNodeID = newNodeID, model_id
					}
					graph.Relation(ctx, session, model_id, newNodeID, []string{edgeLabel})
				}
			}
		}
	}

	return nil
}

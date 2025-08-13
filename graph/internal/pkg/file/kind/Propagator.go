package kind

import (
	"context"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"

	"github.com/boschglobal/dse.schemas/code/go/dse/kind"
	"github.com/boschglobal/dse.sdp/graph/internal/pkg/graph"
)

type PropagatorSpec kind.PropagatorSpec

func newPropagatorspec() *PropagatorSpec {
	return new(PropagatorSpec)
}

func (p *PropagatorSpec) MergeGraph(ctx context.Context, session neo4j.SessionWithContext, kd *KindDoc) error {
	propagator_id := kd.kind_id

	// PROPAGATOR -[HAS]-> SIGNALS
	if p.Signals != nil {
		for _, signal := range *p.Signals {
			matchProps := map[string]string{
				"signal": signal.Signal,
				"target": *signal.Target,
			}
			nodeProps := map[string]any{}
			signal_id, _ := graph.NodeExt(ctx, session, []string{"Sim:Signals"}, matchProps, nodeProps)
			graph.Relation(ctx, session, propagator_id, signal_id, []string{"Has"})

			// SIGNALS -[HAS]-> ENCODING
			if signal.Encoding != nil {
				encodingMatchProps := map[string]string{}
				encodingNodeProps := map[string]any{}
				encoding_id, _ := graph.NodeExt(ctx, session, []string{"Sim:Encoding"}, encodingMatchProps, encodingNodeProps)
				graph.Relation(ctx, session, signal_id, encoding_id, []string{"Has"})

				// ENCODING -[HAS]-> LINEAR
				if signal.Encoding.Linear != nil {
					linearMatchProps := map[string]string{}
					linearNodeProps := map[string]any{
						"factor": signal.Encoding.Linear.Factor,
						"max":    signal.Encoding.Linear.Max,
						"min":    signal.Encoding.Linear.Min,
						"offset": signal.Encoding.Linear.Offset,
					}
					linear_id, _ := graph.NodeExt(ctx, session, []string{"Sim:Linear"}, linearMatchProps, linearNodeProps)
					graph.Relation(ctx, session, encoding_id, linear_id, []string{"Has"})
				}

				// ENCODING -[HAS]-> MAPPING
				if signal.Encoding.Mapping != nil {
					for _, mapping := range *signal.Encoding.Mapping {
						mapMatchProps := map[string]string{
							"name": *mapping.Name,
						}
						mapNodeProps := map[string]any{
							"source": mapping.Source,
							"target": mapping.Target,
						}

						mapping_id, _ := graph.NodeExt(ctx, session, []string{"Sim:Mapping"}, mapMatchProps, mapNodeProps)
						graph.Relation(ctx, session, encoding_id, mapping_id, []string{"Has"})

						// MAPPING -[HAS]-> RANGE
						if signal.Encoding.Mapping != nil {
							RangeMatchProps := map[string]string{}
							RangeNodeProps := map[string]any{
								"min": mapping.Range.Min,
								"max": mapping.Range.Max,
							}
							range_id, _ := graph.NodeExt(ctx, session, []string{"Sim:Range"}, RangeMatchProps, RangeNodeProps)
							graph.Relation(ctx, session, mapping_id, range_id, []string{"Has"})
						}
					}
				}
			}
		}
	}

	// PROPAGATOR -[HAS]-> OPTIONS
	if p.Options != nil {
		if p.Options.Direction != nil {
			matchProps := map[string]string{}
			nodeProps := map[string]any{
				"direction": *p.Options.Direction,
			}
			options_id, _ := graph.NodeExt(ctx, session, []string{"Sim:Options"}, matchProps, nodeProps)
			graph.Relation(ctx, session, propagator_id, options_id, []string{"Has"})
		}
	}

	return nil
}

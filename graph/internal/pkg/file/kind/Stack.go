package kind

import (
	"context"
	"strconv"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"

	"github.com/boschglobal/dse.sdp/graph/internal/pkg/graph"
	"github.com/boschglobal/dse.schemas/code/go/dse/kind"
)

type StackSpec kind.StackSpec

func newStackSpec() *StackSpec {
	return new(StackSpec)
}

func (s *StackSpec) MergeGraph(ctx context.Context, session neo4j.SessionWithContext, kd *KindDoc) error {
	stack_id := kd.kind_id

	// STACK -[HAS]-> MODELINSTANCE
	if s.Models != nil {
		for _, mi := range *s.Models {
			var channelName string
			if mi.Name == "simbus" {
				match_props := map[string]string{
					"name": mi.Name,
				}
				node_props := map[string]any{}
				model_instance_id, _ := graph.NodeExt(ctx, session, []string{"Sim:Simbus"}, match_props, node_props)
				graph.Relation(ctx, session, stack_id, model_instance_id, []string{"Has"})

				// Handle Channels
				if mi.Channels != nil {
					for _, c := range *mi.Channels {
						if c.Name != nil {
							channelName = *c.Name
						}
						channel_match_props := map[string]string{
							"name": channelName,
						}
						channel_node_props := map[string]any{
							"expectedModelCount": c.ExpectedModelCount,
						}
						channel_id, _ := graph.NodeExt(ctx, session, []string{"Sim:SimbusChannel"}, channel_match_props, channel_node_props)
						graph.Relation(ctx, session, model_instance_id, channel_id, []string{"Has"})
					}
				}
			} else if mi.Name != "simbus" {
				match_props := map[string]string{
					"name": mi.Name,
					"uid": strconv.Itoa(mi.Uid),
				}
				node_props := map[string]any{
					"annotations": mi.Annotations,
					"model":       mi.Model.Name,
				}
				model_instance_id, _ := graph.NodeExt(ctx, session, []string{"Sim:ModelInst"}, match_props, node_props)
				graph.Relation(ctx, session, stack_id, model_instance_id, []string{"Has"})

				// Handle Channels and Selectors
				if mi.Channels != nil {
					for _, c := range *mi.Channels {
						if c.Name != nil {
							channelName = *c.Name
						}
						channel_match_props := map[string]string{
							"name": channelName,
						}
						channel_node_props := map[string]any{}
						channel_id, _ := graph.NodeExt(ctx, session, []string{"Sim:Channel"}, channel_match_props, channel_node_props)

						// MODELINST -[ALIAS]-> CHANNEL
						if c.Alias != nil {
							alias_rel_props := map[string]any{
								"name": *c.Alias,
							}
							graph.RelationExt(ctx, session, model_instance_id, channel_id, []string{"Alias"}, alias_rel_props)
						}

						// Handle Selectors
						if c.Selectors != nil && c.Alias != nil {
							for key, value := range *c.Selectors {
								selector_match_props := map[string]string{
									"channelName":  channelName,
									"channelAlias": *c.Alias,
									"selectorName": key,
									"selectorValue": value,
								}
								selector_node_props := map[string]any{}
								selector_id, _ := graph.NodeExt(ctx, session, []string{"Sim:Selector"}, selector_match_props, selector_node_props)
								graph.Relation(ctx, session, model_instance_id, selector_id, []string{"Has"})

								// SELECTOR -[IDENTIFIES]-> CHANNEL
								identifies_rel_props := map[string]any{
									"alias": *c.Alias,
								}
								graph.RelationExt(ctx, session, selector_id, channel_id, []string{"Identifies"}, identifies_rel_props)
							}
						}
					}
				}

				// MODELINST -[HAS]-> RUNTIME
				if mi.Runtime != nil {
					match_props := map[string]string{}
					node_props := map[string]any{
						"env":   mi.Runtime.Env,
						"x32":   mi.Runtime.X32,
						"files": mi.Runtime.Files,
					}
					runtime_id, _ := graph.NodeExt(ctx, session, []string{"Sim:ModelInstanceRuntime"}, match_props, node_props)
					graph.Relation(ctx, session, model_instance_id, runtime_id, []string{"Has"})
				}
			}
		}
	}

	// STACK -[HAS]-> CONNECTION
	if s.Connection != nil {
		if s.Connection.Timeout != nil {
			connection_match_props := map[string]string{
				"timeout": *s.Connection.Timeout,
			}
			connection_node_props := map[string]any{}
			connection_id, _ := graph.NodeExt(ctx, session, []string{"Sim:Connection"}, connection_match_props, connection_node_props)
			graph.Relation(ctx, session, stack_id, connection_id, []string{"Has"})
		}
	}

	// STACK -[HAS]-> RUNTIME
	if s.Runtime != nil {
		match_props := map[string]string{}
		node_props := map[string]any{
			"env":     s.Runtime.Env,
			"stacked": s.Runtime.Stacked,
		}
		runtime_id, _ := graph.NodeExt(ctx, session, []string{"Sim:StackRuntime"}, match_props, node_props)
		graph.Relation(ctx, session, stack_id, runtime_id, []string{"Has"})
	}

	return nil
}

package kind

import (
	"context"
	"log"

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
				}
				node_props := map[string]any{
					"annotations": mi.Annotations,
					"uid":         mi.Uid,
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

				// CREATE INSTANCE_OF RELATIONSHIP
				instance_properties := map[string]any{
					"model_name": mi.Model.Name,
				}

				Query_InstanceOf := `
				    MATCH (inst:ModelInst)
				    MATCH (m:Model{name:$model_name})
				    with inst, m
				    WHERE inst.model = $model_name
				    MERGE (inst)-[r:InstanceOf]->(m)
				`
				_, _ = graph.Query(ctx, session, Query_InstanceOf, instance_properties)

				// CREATE SELECTS RELATIONSHIP
				Query_Selects := `
				MATCH (sg:SignalGroup)-[sgHas:Has]->(l:Label)
				MATCH (mi:ModelInst)-[miHas:Has]->(sl:Selector)
				WHERE sl.selectorName = l.label_name AND sl.selectorValue = l.label_value

				WITH sg, sl, l, mi, COUNT(miHas) AS miCount, COUNT(sgHas) AS sgCount
				WHERE miCount = sgCount
				MERGE (sl)-[:Selects]->(l)
				`
				_, _ = graph.Query(ctx, session, Query_Selects, nil)

				// CREATE ALIAS_OF RELATIONSHIP
				channel_properties := map[string]any{
					"channel_name": channelName,
				}

				query_aliasof := `
				MATCH (sc:SimbusChannel{name:$channel_name})
				MATCH(c:Channel{name:$channel_name})
				with sc, c
				WHERE sc.name = c.name
				MERGE (c)-[a:Belongs]->(sc)
				`
				_, _ = graph.Query(ctx, session, query_aliasof, channel_properties)

				// Run queries to create Represents relationship
				mi_properties := map[string]any{
					"mi_name": mi.Name,
				}

				// First query: Channel Selector Count
				query1 := `
				CALL {
				MATCH (c:Channel)<-[id:Identifies]-(s:Selector)<-[has:Has]-(m:Model)<-[i:InstanceOf]-(mi:ModelInst{name:$mi_name})
				RETURN s, c, id
				UNION
				MATCH (m:Model)<-[i:InstanceOf]-(mi:ModelInst{name:$mi_name})-[h:Has]->(s:Selector)-[id:Identifies]->(c:Channel)
				RETURN s, c, id
				}
				// First query Channel Selector Count
				with c as channel, s as selector
				return channel, count(selector) AS selectorCount
				`

				// Second Query: Channel Label Count
				query2 := `
				CALL {
				MATCH (c:Channel)<-[id:Identifies]-(s:Selector)<-[has:Has]-(m:Model)<-[i:InstanceOf]-(mi:ModelInst{name:$mi_name})
				RETURN s, c, id
				UNION
				MATCH (m:Model)<-[i:InstanceOf]-(mi:ModelInst{name:$mi_name})-[h:Has]->(s:Selector)-[id:Identifies]->(c:Channel)
				RETURN s, c, id
				}
				// Second Query
				with c as channel, s as selector
				MATCH (channel)<-[:Identifies]-(selector)-[selects:Selects]->(l:Label)<-[h:Has]-(sig:SignalGroup)
				RETURN channel, sig, count(l) AS labelCount
				`
				_, _ = graph.Query(ctx, session, query1, mi_properties)
				_, _ = graph.Query(ctx, session, query2, mi_properties)


				// Maps to store selector counts and channel node IDs
				selectorCounts := make(map[int64]int64)
				channelNodeIDs := make(map[int64]int64)

				result1, _ := session.Run(ctx, query1, mi_properties)
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

				result2, _ := session.Run(ctx, query2, mi_properties)
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
									// Create relationship if labelCount >= selectorCount
									if labelCount >= selectorCount {
										relationshipQuery := `
										MATCH (c:Channel) WHERE ID(c) = $channelID
										MATCH (sg:SignalGroup) WHERE ID(sg) = $signalGroupID
										MERGE (c)-[:Represents]->(sg)
										`
										params := map[string]any{
											"channelID":    channelID,
											"signalGroupID": signalGroupID,
										}
										_, err := session.Run(ctx, relationshipQuery, params)
										if err != nil {
											log.Printf("Failed to create relationship: %v", err)
										} else {
											log.Printf("Relationship created: Channel(ID: %d) -> Represents -> SignalGroup(ID: %d)", channelID, signalGroupID)
										}
									} else {
										log.Printf("Skipping channel ID %d (labelCount: %d, selectorCount: %d)", channelID, labelCount, selectorCount)
									}
								}
							}
						}
					}
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

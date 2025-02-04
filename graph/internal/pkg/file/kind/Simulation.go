package kind

import (
	"context"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"

	"github.com/boschglobal/dse.sdp/graph/internal/pkg/graph"
	"github.com/boschglobal/dse.schemas/code/go/dse/ast"
)

type SimulationSpec ast.SimulationSpec

func newSimulationSpec() *SimulationSpec {
	return new(SimulationSpec)
}

func (s *SimulationSpec) MergeGraph(ctx context.Context, session neo4j.SessionWithContext, kd *KindDoc) error {
	simulationID := kd.kind_id

	// SIMULATION -[HAS]-> CHANNELS
	for _, channel := range s.Channels {
		channelMatchProps := map[string]string{"channel_name": channel.Name}
		channelNodeProps := map[string]any{}
		channelID, _ := graph.NodeExt(ctx, session, []string{"Ast:SimulationChannel"}, channelMatchProps, channelNodeProps)
		graph.Relation(ctx, session, simulationID, channelID, []string{"Has"})

		// CHANNEL -[HAS]-> NETWORKS
		if channel.Networks != nil {
			for _, network := range *channel.Networks {
				networkMatchProps := map[string]string{
					"network_name": network.Name,
					"mime_type":    network.MimeType,
				}
				networkNodeProps := map[string]any{}
				networkID, _ := graph.NodeExt(ctx, session, []string{"Ast:Network"}, networkMatchProps, networkNodeProps)
				graph.Relation(ctx, session, channelID, networkID, []string{"Has"})
			}
		}
	}

	// SIMULATION -[HAS]-> STACKS
	for _, stack := range s.Stacks {
		stackMatchProps := map[string]string{"stack_name": stack.Name}
		stackNodeProps := map[string]any{"arch": stack.Arch}
		if stack.Stacked != nil {
			stackNodeProps["stacked"] = *stack.Stacked
		}
		stackID, _ := graph.NodeExt(ctx, session, []string{"Ast:Stack"}, stackMatchProps, stackNodeProps)
		graph.Relation(ctx, session, simulationID, stackID, []string{"Has"})

		// STACKS -[HAS]-> ENV
		if stack.Env != nil {
			for _, env := range *stack.Env {
				envMatchProps := map[string]string{
					"env_name": env.Name,
					"env_value": env.Value,
				}
				envNodeProps := map[string]any{}
				envID, _ := graph.NodeExt(ctx, session, []string{"Ast:Env"}, envMatchProps, envNodeProps)
				graph.Relation(ctx, session, stackID, envID, []string{"Has"})
			}
		}

		// STACK -[HAS]-> MODELINST
		for _, model := range stack.Models {
			modelMatchProps := map[string]string{"model_name": model.Name}
			modelNodeProps := map[string]any{
				"arch": model.Arch,
				"model": model.Model,
			}
			modelID, _ := graph.NodeExt(ctx, session, []string{"Ast:ModelInst"}, modelMatchProps, modelNodeProps)
			graph.Relation(ctx, session, stackID, modelID, []string{"Has"})

			// MODEL -[HAS]-> CHANNELS
			for _, channel := range model.Channels {
				channelMatchProps := map[string]string{"channel_name": channel.Name}
				channelNodeProps := map[string]any{}
				if channel.Alias != "" {
					channelNodeProps["alias"] = channel.Alias
				}
				channelID, _ := graph.NodeExt(ctx, session, []string{"Ast:ModelChannel"}, channelMatchProps, channelNodeProps)
				graph.Relation(ctx, session, modelID, channelID, []string{"Contains"})


				connectQuery := `
				MATCH (mc:Ast:ModelChannel {channel_name: $channelName}),
					  (sc:Ast:SimulationChannel {channel_name: $channelName})
				CREATE (mc)-[:Connects]->(sc)
				`
				params := map[string]any{
					"channelName": channel.Name,
				}
				_, _ = graph.Query(ctx, session, connectQuery, params)
			}

			// MODELINST -[HAS]-> ENV
			if model.Env != nil {
				for _, env := range *model.Env {
					envMatchProps := map[string]string{
						"env_name": env.Name,
						"env_value": env.Value,
					}
					envNodeProps := map[string]any{}
					envID, _ := graph.NodeExt(ctx, session, []string{"Ast:Env"}, envMatchProps, envNodeProps)
					graph.Relation(ctx, session, modelID, envID, []string{"Has"})
				}
			}

			// MODELINST -[HAS]-> WORKFLOWS
			if model.Workflows != nil {
				for _, wf := range *model.Workflows {
					wfMatchProps := map[string]string{
						"workflow_name": wf.Name,
					}
					wfNodeProps := map[string]any{
						"uses": wf.Uses,
					}
					workflowID, _ := graph.NodeExt(ctx, session, []string{"Ast:Workflow"}, wfMatchProps, wfNodeProps)
					graph.Relation(ctx, session, modelID, workflowID, []string{"Has"})

					// WORKFLOW -[HAS]-> VARS
					if wf.Vars != nil {
						for _, vars := range *wf.Vars {
							varMatchProps := map[string]string{
								"var_name":  vars.Name,
								"var_value": vars.Value,
							}
							varNodeProps := map[string]any{}
							varID, _ := graph.NodeExt(ctx, session, []string{"Ast:Var"}, varMatchProps, varNodeProps)
							graph.Relation(ctx, session, workflowID, varID, []string{"Has"})
						}
					}
				}
			}
		}
	}

	// SIMULATION -[HAS]-> USES
	if s.Uses != nil {
		for _, uses := range *s.Uses {
			usesMatchProps := map[string]string{
				"uses_name": uses.Name,
			}
			usesNodeProps := map[string]any{
				"path":    uses.Path,
				"url":     uses.Url,
				"version": uses.Version,
			}
			usesID, _ := graph.NodeExt(ctx, session, []string{"Ast:Uses"}, usesMatchProps, usesNodeProps)
			graph.Relation(ctx, session, simulationID, usesID, []string{"Has"})
		}
	}

	// SIMULATION -[HAS]-> VARS
	if s.Vars != nil {
		for _, vars := range *s.Vars {
			varsMatchProps := map[string]string{
				"var_name":  vars.Name,
				"var_value": vars.Value,
			}
			varsNodeProps := map[string]any{}
			varsID, _ := graph.NodeExt(ctx, session, []string{"Ast:Var"}, varsMatchProps, varsNodeProps)
			graph.Relation(ctx, session, simulationID, varsID, []string{"Has"})
		}
	}

	return nil
}

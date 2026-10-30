package kind

import (
	"context"

	"github.com/boschglobal/dse.schemas/code/go/dse/ast"
	"github.com/boschglobal/dse.sdp/graph/internal/pkg/graph"
)

type SimulationSpec ast.SimulationSpec

const connectChannelQuery = `
	MATCH (mc:Ast:ModelChannel {channel_name: $channelName}),
	      (sc:Ast:SimulationChannel {channel_name: $channelName})
	CREATE (mc)-[:Connects]->(sc)
`

func newSimulationSpec() *SimulationSpec {
	return new(SimulationSpec)
}

func (s *SimulationSpec) MergeGraph(ctx context.Context, session graph.Executor, kd *KindDoc) error {
	w := newWriter(ctx, session)
	id := kd.kindID

	// SIMULATION -[HAS]-> CHANNELS -[HAS]-> NETWORKS
	for _, channel := range s.Channels {
		channelID := w.has(id, "Ast:SimulationChannel", map[string]string{"channel_name": channel.Name}, nil)
		each(channel.Networks, func(n ast.SimulationNetwork) {
			w.has(channelID, "Ast:Network", map[string]string{
				"network_name": n.Name,
				"mime_type":    n.MimeType,
			}, nil)
		})
	}

	// SIMULATION -[HAS]-> STACKS
	for _, stack := range s.Stacks {
		mergeAstStack(w, id, stack)
	}

	// SIMULATION -[HAS]-> USES
	each(s.Uses, func(u ast.Uses) {
		w.has(id, "Ast:Uses",
			map[string]string{"uses_name": u.Name},
			map[string]any{"path": u.Path, "url": u.Url, "version": u.Version})
	})

	// SIMULATION -[HAS]-> VARS
	mergeAstVars(w, id, "Ast:Var", "var", s.Vars)
	return w.err
}

func mergeAstStack(w *writer, simulationID int64, stack ast.Stack) {
	props := map[string]any{"arch": stack.Arch}
	if stack.Stacked != nil {
		props["stacked"] = *stack.Stacked
	}
	stackID := w.has(simulationID, "Ast:Stack", map[string]string{"stack_name": stack.Name}, props)

	// STACK -[HAS]-> ENV
	mergeAstVars(w, stackID, "Ast:Env", "env", stack.Env)

	// STACK -[HAS]-> MODELINST
	for _, model := range stack.Models {
		mergeAstModel(w, stackID, model)
	}
}

func mergeAstModel(w *writer, stackID int64, model ast.Model) {
	modelID := w.has(stackID, "Ast:ModelInst",
		map[string]string{"model_name": model.Name},
		map[string]any{"arch": model.Arch, "model": model.Model})

	// MODELINST -[CONTAINS]-> CHANNELS, connected to the simulation channels.
	for _, channel := range model.Channels {
		props := map[string]any{}
		if channel.Alias != "" {
			props["alias"] = channel.Alias
		}
		w.child(modelID, "Contains", "Ast:ModelChannel", map[string]string{"channel_name": channel.Name}, props)
		w.query(connectChannelQuery, map[string]any{"channelName": channel.Name})
	}

	// MODELINST -[HAS]-> ENV
	mergeAstVars(w, modelID, "Ast:Env", "env", model.Env)

	// MODELINST -[HAS]-> WORKFLOWS -[HAS]-> VARS
	each(model.Workflows, func(wf ast.Workflow) {
		workflowID := w.has(modelID, "Ast:Workflow",
			map[string]string{"workflow_name": wf.Name},
			map[string]any{"uses": wf.Uses})
		mergeAstVars(w, workflowID, "Ast:Var", "var", wf.Vars)
	})
}

// mergeAstVars links name/value variables to parent, keyed as <prefix>_name and <prefix>_value.
func mergeAstVars(w *writer, parent int64, label, prefix string, vars *[]ast.Var) {
	each(vars, func(v ast.Var) {
		w.has(parent, label, map[string]string{
			prefix + "_name":  v.Name,
			prefix + "_value": v.Value,
		}, nil)
	})
}

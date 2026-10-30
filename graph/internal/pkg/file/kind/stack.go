package kind

import (
	"context"
	"strconv"

	schema_kind "github.com/boschglobal/dse.schemas/code/go/dse/kind"
	"github.com/boschglobal/dse.sdp/graph/internal/pkg/graph"
)

type StackSpec schema_kind.StackSpec

func newStackSpec() *StackSpec {
	return new(StackSpec)
}

func (s *StackSpec) MergeGraph(ctx context.Context, session graph.Executor, kd *KindDoc) error {
	w := newWriter(ctx, session)
	id := kd.kindID

	// STACK -[HAS]-> MODELINSTANCE
	each(s.Models, func(mi schema_kind.ModelInstance) {
		if mi.Name == "simbus" {
			mergeSimbus(w, id, mi)
		} else {
			mergeModelInstance(w, id, mi)
		}
	})

	// STACK -[HAS]-> CONNECTION
	if s.Connection != nil && s.Connection.Timeout != nil {
		w.has(id, "Sim:Connection", map[string]string{"timeout": *s.Connection.Timeout}, nil)
	}

	// STACK -[HAS]-> RUNTIME
	if s.Runtime != nil {
		w.has(id, "Sim:StackRuntime", nil, map[string]any{
			"env":     s.Runtime.Env,
			"stacked": s.Runtime.Stacked,
		})
	}
	return w.err
}

func mergeSimbus(w *writer, stackID int64, mi schema_kind.ModelInstance) {
	simbusID := w.has(stackID, "Sim:Simbus", map[string]string{"name": mi.Name}, nil)
	each(mi.Channels, func(c schema_kind.Channel) {
		w.has(simbusID, "Sim:SimbusChannel",
			map[string]string{"name": deref(c.Name)},
			map[string]any{"expectedModelCount": c.ExpectedModelCount})
	})
}

func mergeModelInstance(w *writer, stackID int64, mi schema_kind.ModelInstance) {
	var modelName string
	if mi.Model != nil {
		modelName = mi.Model.Name
	}
	miID := w.has(stackID, "Sim:ModelInst", map[string]string{
		"name": mi.Name,
		"uid":  strconv.Itoa(mi.Uid),
	}, map[string]any{
		"annotations": mi.Annotations,
		"model":       modelName,
	})

	each(mi.Channels, func(c schema_kind.Channel) {
		mergeInstanceChannel(w, miID, c)
	})

	// MODELINST -[HAS]-> RUNTIME
	if mi.Runtime != nil {
		w.has(miID, "Sim:ModelInstanceRuntime", nil, map[string]any{
			"env":   mi.Runtime.Env,
			"x32":   mi.Runtime.X32,
			"files": mi.Runtime.Files,
		})
	}
}

// mergeInstanceChannel creates the channel of a model instance and its selectors.
func mergeInstanceChannel(w *writer, miID int64, c schema_kind.Channel) {
	channelName := deref(c.Name)
	channelID := w.node("Sim:Channel", map[string]string{"name": channelName}, nil)
	if c.Alias == nil {
		return
	}

	// MODELINST -[ALIAS]-> CHANNEL
	w.relate(miID, channelID, "Alias", map[string]any{"name": *c.Alias})

	if c.Selectors == nil {
		return
	}
	for key, value := range *c.Selectors {
		// MODELINST -[HAS]-> SELECTOR -[IDENTIFIES]-> CHANNEL
		selectorID := w.has(miID, "Sim:Selector", map[string]string{
			"channelName":   channelName,
			"channelAlias":  *c.Alias,
			"selectorName":  key,
			"selectorValue": value,
		}, nil)
		w.relate(selectorID, channelID, "Identifies", map[string]any{"alias": *c.Alias})
	}
}

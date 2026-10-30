package kind

import (
	"context"

	schema_kind "github.com/boschglobal/dse.schemas/code/go/dse/kind"
	"github.com/boschglobal/dse.sdp/graph/internal/pkg/graph"
)

type ModelSpec schema_kind.ModelSpec

func newModelSpec() *ModelSpec {
	return new(ModelSpec)
}

func (m *ModelSpec) MergeGraph(ctx context.Context, session graph.Executor, kd *KindDoc) error {
	w := newWriter(ctx, session)
	id := kd.kindID

	// MODEL -[HAS]-> SELECTOR
	each(m.Channels, func(c schema_kind.Channel) {
		if c.Selectors == nil || c.Alias == nil {
			return
		}
		for key, value := range *c.Selectors {
			w.has(id, "Sim:Selector", map[string]string{
				"channelAlias":  *c.Alias,
				"selectorName":  key,
				"selectorValue": value,
			}, nil)
		}
	})

	if m.Runtime != nil {
		m.mergeRuntime(w, id)
	}

	mergeAnnotationGraph(w, kd, id)
	return w.err
}

// mergeRuntime creates MODEL -[HAS]-> RUNTIME and its libraries.
func (m *ModelSpec) mergeRuntime(w *writer, modelID int64) {
	runtime := m.Runtime
	runtimeID := w.has(modelID, "Sim:Runtime", nil, nil)

	each(runtime.Dynlib, func(d schema_kind.LibrarySpec) {
		w.has(runtimeID, "Sim:Dynlib", map[string]string{
			"arch": deref(d.Arch),
			"os":   deref(d.Os),
			"path": d.Path,
		}, map[string]any{
			"annotations": d.Annotations,
			"libs":        d.Libs,
			"variant":     d.Variant,
		})
	})

	each(runtime.Executable, func(e schema_kind.ExecutableSpec) {
		w.has(runtimeID, "Sim:Executable", map[string]string{
			"arch": deref(e.Arch),
			"os":   deref(e.Os),
		}, map[string]any{
			"annotations": e.Annotations,
			"libs":        e.Libs,
		})
	})

	if runtime.Gateway != nil {
		w.has(runtimeID, "Sim:Gateway", nil, map[string]any{
			"annotations": runtime.Gateway.Annotations,
		})
	}

	each(runtime.Mcl, func(l schema_kind.LibrarySpec) {
		w.has(runtimeID, "Sim:Mcl", map[string]string{
			"arch": deref(l.Arch),
			"os":   deref(l.Os),
			"path": l.Path,
		}, map[string]any{
			"annotations": l.Annotations,
			"libs":        l.Libs,
			"variant":     l.Variant,
		})
	})
}

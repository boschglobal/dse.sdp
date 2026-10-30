package kind

import (
	"context"

	schema_kind "github.com/boschglobal/dse.schemas/code/go/dse/kind"
	"github.com/boschglobal/dse.sdp/graph/internal/pkg/graph"
)

type PropagatorSpec schema_kind.PropagatorSpec

func newPropagatorSpec() *PropagatorSpec {
	return new(PropagatorSpec)
}

func (p *PropagatorSpec) MergeGraph(ctx context.Context, session graph.Executor, kd *KindDoc) error {
	w := newWriter(ctx, session)
	id := kd.kindID

	// PROPAGATOR -[HAS]-> SIGNALS
	each(p.Signals, func(signal schema_kind.SignalEncoding) {
		signalID := w.has(id, "Sim:Signals", map[string]string{
			"signal": signal.Signal,
			"target": deref(signal.Target),
		}, nil)
		if signal.Encoding != nil {
			mergeEncoding(w, signalID, signal)
		}
	})

	// PROPAGATOR -[HAS]-> OPTIONS
	if p.Options != nil && p.Options.Direction != nil {
		w.has(id, "Sim:Options", nil, map[string]any{"direction": *p.Options.Direction})
	}
	return w.err
}

// mergeEncoding creates SIGNALS -[HAS]-> ENCODING with its linear and mapping nodes.
func mergeEncoding(w *writer, signalID int64, signal schema_kind.SignalEncoding) {
	encoding := signal.Encoding
	encodingID := w.has(signalID, "Sim:Encoding", nil, nil)

	if l := encoding.Linear; l != nil {
		w.has(encodingID, "Sim:Linear", nil, map[string]any{
			"factor": l.Factor,
			"max":    l.Max,
			"min":    l.Min,
			"offset": l.Offset,
		})
	}

	if encoding.Mapping == nil {
		return
	}
	for _, mapping := range *encoding.Mapping {
		mappingID := w.has(encodingID, "Sim:Mapping",
			map[string]string{"name": deref(mapping.Name)},
			map[string]any{"source": mapping.Source, "target": mapping.Target})
		if mapping.Range != nil {
			w.has(mappingID, "Sim:Range", nil, map[string]any{
				"min": mapping.Range.Min,
				"max": mapping.Range.Max,
			})
		}
	}
}

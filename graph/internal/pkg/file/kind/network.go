package kind

import (
	"context"

	schema_kind "github.com/boschglobal/dse.schemas/code/go/dse/kind"
	"github.com/boschglobal/dse.sdp/graph/internal/pkg/graph"
)

type NetworkSpec schema_kind.NetworkSpec

func newNetworkSpec() *NetworkSpec {
	return new(NetworkSpec)
}

func (n *NetworkSpec) MergeGraph(ctx context.Context, session graph.Executor, kd *KindDoc) error {
	w := newWriter(ctx, session)
	if n.Messages == nil {
		return nil
	}

	// NETWORK -[HAS]-> MESSAGES
	for _, m := range *n.Messages {
		messageID := w.has(kd.kindID, "Sim:Messages",
			map[string]string{"message": m.Message},
			map[string]any{"annotations": m.Annotations})

		// MESSAGES -[HAS]-> SIGNALS
		if m.Signals != nil {
			for _, signal := range *m.Signals {
				w.has(messageID, "Sim:Signals",
					map[string]string{"signal": signal.Signal},
					map[string]any{"annotations": signal.Annotations})
			}
		}

		// MESSAGES -[HAS]-> FUNCTIONS -[HAS]-> {ENCODE, DECODE}
		if m.Functions != nil {
			functionsID := w.has(messageID, "Sim:Functions", nil, nil)
			if m.Functions.Decode != nil {
				for _, f := range *m.Functions.Decode {
					w.mergeFunction(functionsID, "Sim:Decode", f.Function, f.Annotations)
				}
			}
			if m.Functions.Encode != nil {
				for _, f := range *m.Functions.Encode {
					w.mergeFunction(functionsID, "Sim:Encode", f.Function, f.Annotations)
				}
			}
		}
	}
	return w.err
}

// mergeFunction links a named function node; unnamed functions are skipped.
func (w *writer) mergeFunction(parent int64, label string, function *string, annotations any) {
	if function == nil {
		return
	}
	w.has(parent, label,
		map[string]string{"function": *function},
		map[string]any{"annotations": annotations})
}

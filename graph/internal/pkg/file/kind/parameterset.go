package kind

import (
	"context"

	schema_kind "github.com/boschglobal/dse.schemas/code/go/dse/kind"
	"github.com/boschglobal/dse.sdp/graph/internal/pkg/graph"
)

type ParameterSetSpec schema_kind.ParameterSetSpec

func newParameterSetSpec() *ParameterSetSpec {
	return new(ParameterSetSpec)
}

func (ps *ParameterSetSpec) MergeGraph(ctx context.Context, session graph.Executor, kd *KindDoc) error {
	w := newWriter(ctx, session)

	// PARAMETERSET -[HAS]-> PARAMETERS
	for _, p := range ps.Parameters {
		w.has(kd.kindID, "Sim:Parameters",
			map[string]string{"parameter": p.Parameter},
			map[string]any{"annotations": p.Annotations, "value": p.Value})
	}
	return w.err
}

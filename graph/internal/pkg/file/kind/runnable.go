package kind

import (
	"context"

	schema_kind "github.com/boschglobal/dse.schemas/code/go/dse/kind"
	"github.com/boschglobal/dse.sdp/graph/internal/pkg/graph"
)

type RunnableSpec schema_kind.RunnableSpec

func newRunnableSpec() *RunnableSpec {
	return new(RunnableSpec)
}

func (rn *RunnableSpec) MergeGraph(ctx context.Context, session graph.Executor, kd *KindDoc) error {
	w := newWriter(ctx, session)

	// RUNNABLE -[HAS]-> TASKS
	for _, t := range rn.Tasks {
		w.has(kd.kindID, "Sim:Tasks",
			map[string]string{"function": t.Function},
			map[string]any{"schedule": t.Schedule})
	}
	return w.err
}

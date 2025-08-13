package kind

import (
	"context"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"

	"github.com/boschglobal/dse.schemas/code/go/dse/kind"
	"github.com/boschglobal/dse.sdp/graph/internal/pkg/graph"
)

type RunnableSpec kind.RunnableSpec

func newRunnableSpec() *RunnableSpec {
	return new(RunnableSpec)
}

func (rn *RunnableSpec) MergeGraph(ctx context.Context, session neo4j.SessionWithContext, kd *KindDoc) error {
	runnable_id := kd.kind_id

	// RUNNABLE -[HAS]-> TASKS
	for _, t := range rn.Tasks {
		match_props := map[string]string{
			"function": t.Function,
		}
		node_props := map[string]any{
			"schedule": t.Schedule,
		}
		task_id, _ := graph.NodeExt(ctx, session, []string{"Sim:Tasks"}, match_props, node_props)
		graph.Relation(ctx, session, runnable_id, task_id, []string{"Has"})
	}

	return nil
}

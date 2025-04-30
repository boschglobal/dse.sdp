package kind

import (
	"context"
	"fmt"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"

	"github.com/boschglobal/dse.sdp/graph/internal/pkg/graph"
	"github.com/boschglobal/dse.schemas/code/go/dse/kind"
)

type ParameterSetSpec kind.ParameterSetSpec

func newParameterSetSpec() *ParameterSetSpec {
	return new(ParameterSetSpec)
}

func (ps *ParameterSetSpec) MergeGraph(ctx context.Context, session neo4j.SessionWithContext, kd *KindDoc) error {
    parametersetID := kd.kind_id

	// PARAMETERSET -[HAS]-> PARAMETERS
    for _, p := range ps.Parameters {
        matchProps := map[string]string{
            "parameter": p.Parameter,
        }
        nodeProps := map[string]interface{}{
            "annotations": p.Annotations,
			"value": p.Value,
        }
        parametersID, err := graph.NodeExt(ctx, session, []string{"Sim:Parameters"}, matchProps, nodeProps)
        if err != nil {
            return fmt.Errorf("failed to create or retrieve message node: %w", err)
        }
        graph.Relation(ctx, session, parametersetID, parametersID, []string{"Has"})
    }

	return nil
}

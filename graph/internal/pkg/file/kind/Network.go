package kind

import (
	"context"
    "fmt"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"

	"github.com/boschglobal/dse.sdp/graph/internal/pkg/graph"
	"github.com/boschglobal/dse.schemas/code/go/dse/kind"
)

type NetworkSpec kind.NetworkSpec

func newNetworkSpec() *NetworkSpec {
	return new(NetworkSpec)
}

func (n *NetworkSpec) MergeGraph(ctx context.Context, session neo4j.SessionWithContext, kd *KindDoc) error {
    networkID := kd.kind_id

    // NETWORK -[HAS]-> MESSAGES
    for _, m := range n.Messages {
        matchProps := map[string]string{
            "message": m.Message,
        }
        nodeProps := map[string]interface{}{
            "annotations": m.Annotations,
        }
        messageID, err := graph.NodeExt(ctx, session, []string{"Sim:Messages"}, matchProps, nodeProps)
        if err != nil {
            return fmt.Errorf("failed to create or retrieve message node: %w", err)
        }
        graph.Relation(ctx, session, networkID, messageID, []string{"Has"})

        // MESSAGES -[HAS]-> SIGNALS
        if m.Signals != nil {
            for _, signal := range *m.Signals {
                signalMatchProps := map[string]string{
                    "signal": signal.Signal,
                }
                signalNodeProps := map[string]interface{}{
                    "annotations": signal.Annotations,
                }
                signalID, err := graph.NodeExt(ctx, session, []string{"Sim:Signals"}, signalMatchProps, signalNodeProps)
                if err != nil {
                    return fmt.Errorf("failed to create or retrieve signal node: %w", err)
                }
                graph.Relation(ctx, session, messageID, signalID, []string{"Has"})
            }
        }

        // MESSAGES -[HAS]-> FUNCTIONS -[HAS]-> {ENCODE, DECODE}
        if m.Functions != nil {
            functionsMatchProps := map[string]string{}
            functionsNodeProps := map[string]any{}
            functionsID, _ := graph.NodeExt(ctx, session, []string{"Sim:Functions"}, functionsMatchProps, functionsNodeProps)
            graph.Relation(ctx, session, messageID, functionsID, []string{"Has"})

            if m.Functions.Decode != nil {
                for _, decodeFunction := range *m.Functions.Decode {
                    decodeMatchProps := map[string]string{
                        "function": decodeFunction.Function,
                    }
                    decodeNodeProps := map[string]interface{}{
                        "annotations": decodeFunction.Annotations,
                    }
                    decodeFuncID, _ := graph.NodeExt(ctx, session, []string{"Sim:Decode"}, decodeMatchProps, decodeNodeProps)
                    graph.Relation(ctx, session, functionsID, decodeFuncID, []string{"Has"})
                }
            }

            if m.Functions.Encode != nil {
                for _, encodeFunction := range *m.Functions.Encode {
                    encodeMatchProps := map[string]string{
                        "function": encodeFunction.Function,
                    }
                    encodeNodeProps := map[string]interface{}{
                        "annotations": encodeFunction.Annotations,
                    }
                    encodeFuncID, _ := graph.NodeExt(ctx, session, []string{"Sim:Encode"}, encodeMatchProps, encodeNodeProps)
                    graph.Relation(ctx, session, functionsID, encodeFuncID, []string{"Has"})
                }
            }
        }
    }

    return nil
}

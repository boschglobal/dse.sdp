package kind

import (
	"context"
	"strconv"

	schema_kind "github.com/boschglobal/dse.schemas/code/go/dse/kind"
	"github.com/boschglobal/dse.sdp/graph/internal/pkg/graph"
)

type SignalGroupSpec schema_kind.SignalGroupSpec

const mergeLabelQuery = `
	MATCH (sg) WHERE id(sg) = $signalgroup_id
	MERGE (sg)-[:Has]->(l:Sim:Label {label_name: $label_name})
	ON CREATE SET l += $props
	ON MATCH SET l += $props
	RETURN id(l) AS id
`

func newSignalGroupSpec() *SignalGroupSpec {
	return new(SignalGroupSpec)
}

func (sg *SignalGroupSpec) MergeGraph(ctx context.Context, session graph.Executor, kd *KindDoc) error {
	w := newWriter(ctx, session)
	id := kd.kindID

	// SIGNALGROUP -[CONTAINS]-> SIGNAL
	for _, s := range sg.Signals {
		props := map[string]any{}
		if s.Annotations != nil {
			props["annotations"] = *s.Annotations
		}
		w.child(id, "Contains", "Sim:Signal", map[string]string{
			"name":            s.Signal,
			"signal_group_id": strconv.FormatInt(id, 10),
		}, props)
	}

	// SIGNALGROUP -[HAS]-> LABEL
	for name, value := range kd.Metadata.Labels {
		w.query(mergeLabelQuery, map[string]any{
			"signalgroup_id": id,
			"label_name":     name,
			"label_value":    value,
			"props":          map[string]any{"label_value": value},
		})
	}

	mergeAnnotationGraph(w, kd, id)
	return w.err
}

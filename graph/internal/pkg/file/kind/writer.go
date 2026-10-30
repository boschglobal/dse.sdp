package kind

import (
	"context"

	"github.com/boschglobal/dse.sdp/graph/internal/pkg/graph"
)

// writer builds a graph fragment. It remembers the first error and turns later
// calls into no-ops, so spec code can be written without per-call error checks.
type writer struct {
	ctx     context.Context
	session graph.Executor
	err     error
}

func newWriter(ctx context.Context, session graph.Executor) *writer {
	return &writer{ctx: ctx, session: session}
}

func (w *writer) check(err error) {
	if w.err == nil {
		w.err = err
	}
}

// node merges a node, returning its id (-1 after a failure).
func (w *writer) node(label string, match map[string]string, props map[string]any) int64 {
	if w.err != nil {
		return -1
	}
	if match == nil {
		match = map[string]string{}
	}
	if props == nil {
		props = map[string]any{}
	}
	id, err := graph.NodeExt(w.ctx, w.session, []string{label}, match, props)
	w.check(err)
	return id
}

// relate merges a relation between two nodes.
func (w *writer) relate(start, end int64, rel string, props map[string]any) {
	if w.err != nil {
		return
	}
	var err error
	if props == nil {
		_, err = graph.Relation(w.ctx, w.session, start, end, []string{rel})
	} else {
		_, err = graph.RelationExt(w.ctx, w.session, start, end, []string{rel}, props)
	}
	w.check(err)
}

// child merges a node and links it from parent with the given relation.
func (w *writer) child(parent int64, rel, label string, match map[string]string, props map[string]any) int64 {
	id := w.node(label, match, props)
	w.relate(parent, id, rel, nil)
	return id
}

// has is child with a "Has" relation.
func (w *writer) has(parent int64, label string, match map[string]string, props map[string]any) int64 {
	return w.child(parent, "Has", label, match, props)
}

func (w *writer) query(query string, params map[string]any) {
	if w.err != nil {
		return
	}
	_, err := graph.Query(w.ctx, w.session, query, params)
	w.check(err)
}

func deref[T any](p *T) T {
	if p == nil {
		var zero T
		return zero
	}
	return *p
}

func each[T any](s *[]T, fn func(T)) {
	if s == nil {
		return
	}
	for _, v := range *s {
		fn(v)
	}
}

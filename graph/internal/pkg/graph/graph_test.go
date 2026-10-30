package graph_test

import (
	"bytes"
	"context"
	"errors"
	"log/slog"
	"testing"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/boschglobal/dse.sdp/graph/internal/pkg/graph"
	"github.com/boschglobal/dse.sdp/graph/internal/pkg/graph/graphtest"
)

func TestMain(m *testing.M) {
	graphtest.Main(m)
}

func count(t *testing.T, ctx context.Context, session neo4j.SessionWithContext, match string, params map[string]any) int64 {
	t.Helper()
	return graphtest.Count(t, ctx, session, match+" RETURN count(*) AS n", params)
}

func TestSessionWithoutDriver(t *testing.T) {
	_, err := graph.Session(context.Background())
	assert.Error(t, err)
	assert.NotPanics(t, func() { graph.Close(context.Background()) })
}

func TestConnect(t *testing.T) {
	ctx, session := graphtest.Connect(t)
	assert.NotNil(t, session)
	assert.Equal(t, int64(1), graphtest.Count(t, ctx, session, "RETURN 1 AS n", nil))
}

func TestNode(t *testing.T) {
	ctx, session := graphtest.Connect(t)

	id, err := graph.Node(ctx, session, []string{"Sim:Model", "Extra"}, "foo")
	require.NoError(t, err)

	// Merging again returns the same node.
	again, err := graph.Node(ctx, session, []string{"Sim:Model", "Extra"}, "foo")
	require.NoError(t, err)
	assert.Equal(t, id, again)

	other, err := graph.Node(ctx, session, []string{"Sim:Model", "Extra"}, "bar")
	require.NoError(t, err)
	assert.NotEqual(t, id, other)

	assert.Equal(t, int64(2), count(t, ctx, session, "MATCH (n:Sim:Model:Extra)", nil))
	assert.Equal(t, int64(1), count(t, ctx, session, "MATCH (n:Sim:Model:Extra {name: 'foo'})", nil))
}

func TestNodeExt(t *testing.T) {
	ctx, session := graphtest.Connect(t)
	labels := []string{"Sim:Thing"}
	match := map[string]string{"name": "foo", "uid": "1"}

	id, err := graph.NodeExt(ctx, session, labels, match, map[string]any{"colour": "red", "size": 1})
	require.NoError(t, err)

	// Same match values update the properties of the existing node.
	again, err := graph.NodeExt(ctx, session, labels, match, map[string]any{"colour": "blue"})
	require.NoError(t, err)
	assert.Equal(t, id, again)

	records := graphtest.Cypher(t, ctx, session, "MATCH (n:Sim:Thing) RETURN n.colour AS colour, n.size AS size", nil)
	require.Len(t, records, 1)
	colour, _ := records[0].Get("colour")
	size, _ := records[0].Get("size")
	assert.Equal(t, "blue", colour)
	assert.EqualValues(t, 1, size)

	// A different match value creates a new node.
	other, err := graph.NodeExt(ctx, session, labels, map[string]string{"name": "foo", "uid": "2"}, nil)
	require.NoError(t, err)
	assert.NotEqual(t, id, other)
	assert.Equal(t, int64(2), count(t, ctx, session, "MATCH (n:Sim:Thing)", nil))
}

func TestNodeExtNoMatch(t *testing.T) {
	ctx, session := graphtest.Connect(t)

	id, err := graph.NodeExt(ctx, session, []string{"Sim:Singleton"}, map[string]string{}, map[string]any{"a": 1})
	require.NoError(t, err)
	again, err := graph.NodeExt(ctx, session, []string{"Sim:Singleton"}, map[string]string{}, map[string]any{"a": 2})
	require.NoError(t, err)
	assert.Equal(t, id, again)
}

func TestNodeExtQuoting(t *testing.T) {
	ctx, session := graphtest.Connect(t)
	_, err := graph.Node(ctx, session, []string{"Keep"}, "keep")
	require.NoError(t, err)

	hostile := []string{
		"O'Brien",
		"x'}) DETACH DELETE (k) //",
		"back`tick",
		`quote"double`,
	}
	for _, value := range hostile {
		_, err := graph.NodeExt(ctx, session, []string{"Sim:Hostile"}, map[string]string{"name": value}, nil)
		require.NoError(t, err, value)
		assert.Equal(t, int64(1), count(t, ctx, session, "MATCH (n:Sim:Hostile {name: $name})", map[string]any{"name": value}), value)
	}
	assert.Equal(t, int64(1), count(t, ctx, session, "MATCH (n:Keep)", nil))
}

func TestNodeExtHostileKeysAndLabels(t *testing.T) {
	ctx, session := graphtest.Connect(t)

	_, err := graph.NodeExt(ctx, session, []string{"Sim:We`ird"}, map[string]string{"key with space": "v"}, nil)
	require.NoError(t, err)
	assert.Equal(t, int64(1), count(t, ctx, session, "MATCH (n:Sim:`We``ird` {`key with space`: 'v'})", nil))
}

func TestRelation(t *testing.T) {
	ctx, session := graphtest.Connect(t)
	a, _ := graph.Node(ctx, session, []string{"A"}, "a")
	b, _ := graph.Node(ctx, session, []string{"B"}, "b")

	id, err := graph.Relation(ctx, session, a, b, []string{"Has"})
	require.NoError(t, err)
	assert.GreaterOrEqual(t, id, int64(0))

	// Merging again returns the same relation.
	again, err := graph.Relation(ctx, session, a, b, []string{"Has"})
	require.NoError(t, err)
	assert.Equal(t, id, again)

	assert.Equal(t, int64(1), count(t, ctx, session, "MATCH (:A)-[r:Has]->(:B)", nil))
	assert.Equal(t, int64(0), count(t, ctx, session, "MATCH (:B)-[r:Has]->(:A)", nil))
}

func TestRelationMissingNode(t *testing.T) {
	ctx, session := graphtest.Connect(t)
	a, _ := graph.Node(ctx, session, []string{"A"}, "a")

	id, err := graph.Relation(ctx, session, a, a+1000, []string{"Has"})
	assert.Error(t, err)
	assert.Equal(t, int64(-1), id)
}

func TestRelationExt(t *testing.T) {
	ctx, session := graphtest.Connect(t)
	a, _ := graph.Node(ctx, session, []string{"A"}, "a")
	b, _ := graph.Node(ctx, session, []string{"B"}, "b")

	id, err := graph.RelationExt(ctx, session, a, b, []string{"Alias"}, map[string]any{"name": "x"})
	require.NoError(t, err)
	again, err := graph.RelationExt(ctx, session, a, b, []string{"Alias"}, map[string]any{"name": "y"})
	require.NoError(t, err)
	assert.Equal(t, id, again)

	assert.Equal(t, int64(1), count(t, ctx, session, "MATCH (:A)-[r:Alias {name: 'y'}]->(:B)", nil))
	assert.Equal(t, int64(1), count(t, ctx, session, "MATCH ()-[r:Alias]->()", nil))
}

func TestQuery(t *testing.T) {
	ctx, session := graphtest.Connect(t)

	id, err := graph.Query(ctx, session, "CREATE (n:Q {v: $v}) RETURN id(n) AS id", map[string]any{"v": 7})
	require.NoError(t, err)
	assert.GreaterOrEqual(t, id, int64(0))

	// A query without a result row is not an error.
	id, err = graph.Query(ctx, session, "MATCH (n:Missing) RETURN id(n) AS id", nil)
	assert.NoError(t, err)
	assert.Equal(t, int64(-1), id)

	id, err = graph.Query(ctx, session, "THIS IS NOT CYPHER", nil)
	assert.Error(t, err)
	assert.Equal(t, int64(-1), id)
}

func TestQueryRecord(t *testing.T) {
	ctx, session := graphtest.Connect(t)

	record, err := graph.QueryRecord(ctx, session, "RETURN $v AS v", map[string]any{"v": "hello"})
	require.NoError(t, err)
	require.NotNil(t, record)
	v, _ := record.Get("v")
	assert.Equal(t, "hello", v)

	// No single record yields nil without an error (and without panicking).
	record, err = graph.QueryRecord(ctx, session, "MATCH (n:Missing) RETURN n", nil)
	assert.NoError(t, err)
	assert.Nil(t, record)

	record, err = graph.QueryRecord(ctx, session, "THIS IS NOT CYPHER", nil)
	assert.Error(t, err)
	assert.Nil(t, record)
}

func TestDrop(t *testing.T) {
	tests := []struct {
		option    string
		remaining map[string]int64
		wantErr   bool
	}{
		{"ast", map[string]int64{"Ast": 0, "Sim": 1, "Other": 1}, false},
		{"sim", map[string]int64{"Ast": 1, "Sim": 0, "Other": 1}, false},
		{"--all", map[string]int64{"Ast": 0, "Sim": 0, "Other": 0}, false},
		{"bogus", map[string]int64{"Ast": 1, "Sim": 1, "Other": 1}, true},
	}
	for _, tt := range tests {
		t.Run(tt.option, func(t *testing.T) {
			ctx, session := graphtest.Connect(t)
			for _, label := range []string{"Ast", "Sim", "Other"} {
				_, err := graph.Node(ctx, session, []string{label}, "n")
				require.NoError(t, err)
			}

			err := graph.Drop(ctx, tt.option)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			for label, want := range tt.remaining {
				assert.Equal(t, want, count(t, ctx, session, "MATCH (n:"+label+")", nil), label)
			}
		})
	}
}

func TestDropDetachesRelations(t *testing.T) {
	ctx, session := graphtest.Connect(t)
	a, _ := graph.Node(ctx, session, []string{"Ast"}, "a")
	b, _ := graph.Node(ctx, session, []string{"Sim"}, "b")
	_, err := graph.Relation(ctx, session, a, b, []string{"Has"})
	require.NoError(t, err)

	require.NoError(t, graph.Drop(ctx, "ast"))
	assert.Equal(t, int64(0), count(t, ctx, session, "MATCH ()-[r]->()", nil))
	assert.Equal(t, int64(1), count(t, ctx, session, "MATCH (n:Sim)", nil))
}

func TestWriteTxCommit(t *testing.T) {
	ctx, session := graphtest.Connect(t)

	err := graph.WriteTx(ctx, session, func(ex graph.Executor) error {
		a, err := graph.Node(ctx, ex, []string{"A"}, "a")
		if err != nil {
			return err
		}
		b, err := graph.Node(ctx, ex, []string{"B"}, "b")
		if err != nil {
			return err
		}
		_, err = graph.Relation(ctx, ex, a, b, []string{"Has"})
		return err
	})
	require.NoError(t, err)
	assert.Equal(t, int64(1), count(t, ctx, session, "MATCH (:A)-[:Has]->(:B)", nil))
}

func TestWriteTxRollback(t *testing.T) {
	ctx, session := graphtest.Connect(t)
	failure := errors.New("stop")

	err := graph.WriteTx(ctx, session, func(ex graph.Executor) error {
		if _, err := graph.Node(ctx, ex, []string{"A"}, "a"); err != nil {
			return err
		}
		return failure
	})
	assert.ErrorIs(t, err, failure)
	assert.Equal(t, int64(0), count(t, ctx, session, "MATCH (n)", nil))
}

func TestBoltTracing(t *testing.T) {
	ctx, session := graphtest.Connect(t)

	var buf bytes.Buffer
	previous := slog.Default()
	slog.SetDefault(slog.New(slog.NewTextHandler(&buf, &slog.HandlerOptions{Level: graph.LevelTrace})))
	defer slog.SetDefault(previous)

	_, err := graph.Node(ctx, session, []string{"Traced"}, "marker")
	require.NoError(t, err)

	out := buf.String()
	assert.Contains(t, out, "MERGE")
	assert.Contains(t, out, "marker")
	assert.Contains(t, out, "RECORD")
}

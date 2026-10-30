package graph

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/boschglobal/dse.sdp/graph/internal/pkg/graph/graphtest"
)

func TestMain(m *testing.M) {
	graphtest.Main(m)
}

const createFollows = "CREATE (foo {name: 'Foo'})-[r:FOLLOWS]->(bar {name: 'Bar'})"

func TestGraphDrop(t *testing.T) {
	uri := graphtest.URI(t)
	ctx, session := graphtest.Connect(t)
	graphtest.Cypher(t, ctx, session, createFollows, nil)

	cmd := NewGraphDropCommand("drop")
	require.NoError(t, cmd.Parse([]string{"--db", uri, "--all"}))
	require.NoError(t, cmd.Run())

	assert.Equal(t, int64(0), graphtest.Count(t, ctx, session, "MATCH (n) RETURN count(*) AS n", nil))
}

func TestGraphDropInvalidUsage(t *testing.T) {
	uri := graphtest.URI(t)
	ctx, session := graphtest.Connect(t)
	graphtest.Cypher(t, ctx, session, createFollows, nil)

	for _, args := range [][]string{{"--db", uri}, {"--db", uri, "bogus"}} {
		cmd := NewGraphDropCommand("drop")
		require.NoError(t, cmd.Parse(args))
		assert.Error(t, cmd.Run(), args)
	}
	assert.Equal(t, int64(2), graphtest.Count(t, ctx, session, "MATCH (n) RETURN count(*) AS n", nil))
}

func TestGraphExport(t *testing.T) {
	uri := graphtest.URI(t)
	ctx, session := graphtest.Connect(t)
	graphtest.Cypher(t, ctx, session, createFollows, nil)
	exportFile := filepath.Join(t.TempDir(), "export.cyp")

	cmd := NewGraphExportCommand("export")
	require.NoError(t, cmd.Parse([]string{"--db", uri, exportFile}))
	require.NoError(t, cmd.Run())

	content, err := os.ReadFile(exportFile)
	require.NoError(t, err)
	assert.Contains(t, string(content), "CREATE (n:_IMPORT_ID {name: 'Foo',")
	assert.Contains(t, string(content), "CREATE (n:_IMPORT_ID {name: 'Bar',")
	assert.Contains(t, string(content), "CREATE (n)-[:FOLLOWS {}]->(m)")
}

func TestGraphExportWithoutFile(t *testing.T) {
	cmd := NewGraphExportCommand("export")
	require.NoError(t, cmd.Parse([]string{"--db", "bolt://127.0.0.1:1"}))
	assert.Error(t, cmd.Run())
}

func TestGraphImport(t *testing.T) {
	uri := graphtest.URI(t)
	ctx, session := graphtest.Connect(t)

	cmd := NewGraphImportCommand("import")
	require.NoError(t, cmd.Parse([]string{"--db", uri, "../../pkg/file/kind/testdata/yaml"}))
	require.NoError(t, cmd.Run())

	assert.Equal(t, int64(1), graphtest.Count(t, ctx, session,
		"MATCH (:File {name: 'stack.yaml'})-[:Contains]->(:Sim:Stack {name: 'stack'}) RETURN count(*) AS n", nil))
	assert.Equal(t, int64(1), graphtest.Count(t, ctx, session,
		"MATCH (:File {name: 'runnable.yaml'})-[:Contains]->(:Sim:Runnable {name: 'target'}) RETURN count(*) AS n", nil))
}

func TestGraphImportErrors(t *testing.T) {
	uri := graphtest.URI(t)

	cmd := NewGraphImportCommand("import")
	require.NoError(t, cmd.Parse([]string{"--db", uri}))
	assert.Error(t, cmd.Run(), "no path")

	cmd = NewGraphImportCommand("import")
	require.NoError(t, cmd.Parse([]string{"--db", uri, filepath.Join(t.TempDir(), "absent")}))
	assert.Error(t, cmd.Run(), "missing path")

	cmd = NewGraphImportCommand("import")
	require.NoError(t, cmd.Parse([]string{"--db", uri, t.TempDir()}))
	assert.Error(t, cmd.Run(), "no yaml files")
}

func TestGraphPing(t *testing.T) {
	uri := graphtest.URI(t)

	cmd := NewGraphPingCommand("ping")
	require.NoError(t, cmd.Parse([]string{"--db", uri, "--retry", "1"}))
	assert.NoError(t, cmd.Run())
}

func TestGraphPingUnreachable(t *testing.T) {
	cmd := NewGraphPingCommand("ping")
	require.NoError(t, cmd.Parse([]string{"--db", "bolt://127.0.0.1:1", "--retry", "1"}))
	assert.Error(t, cmd.Run())
}

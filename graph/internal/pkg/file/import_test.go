package file

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/boschglobal/dse.sdp/graph/internal/pkg/graph/graphtest"
)

func TestMain(m *testing.M) {
	graphtest.Main(m)
}

const yamlDir = "kind/testdata/yaml"

// Each case imports a fixture and checks graph shapes: query -> expected count.
var importTests = []struct {
	file   string
	counts map[string]int64
}{
	{"signalgroup.yaml", map[string]int64{
		"MATCH (:File {name: 'signalgroup.yaml'})-[:Contains]->(:Sim:SignalGroup {signalgroup_name: 'E2M_M2E'})": 1,
		"MATCH (:Sim:SignalGroup)-[:Contains]->(:Sim:Signal)":                                                    2,
		"MATCH (:Sim:SignalGroup)-[:Has]->(:Sim:Label {label_name: 'channel', label_value: 'E2M_M2E'})":          1,
	}},
	{"model.yaml", map[string]int64{
		"MATCH (:File {name: 'model.yaml'})-[:Contains]->(:Sim:Model {name: 'Network'})":                              1,
		"MATCH (:Sim:Model)-[:Has]->(:Sim:Runtime)-[:Has]->(:Sim:Dynlib {os: 'linux', arch: 'amd64'})":                1,
		"MATCH (:Sim:Model)-[:Has]->(:Sim:Selector {channelAlias: 'signal_channel', selectorValue: 'signal_vector'})": 1,
		"MATCH (:Sim:Model)-[:Has]->(:Sim:Selector)":                                                                  2,
	}},
	{"stack.yaml", map[string]int64{
		"MATCH (:File {name: 'stack.yaml'})-[:Contains]->(:Sim:Stack {name: 'stack'})":                 1,
		"MATCH (:Sim:Stack)-[:Has]->(:Sim:Simbus {name: 'simbus'})-[:Has]->(:Sim:SimbusChannel)":       2,
		"MATCH (:Sim:Stack)-[:Has]->(:Sim:ModelInst {name: 'stub_inst', uid: '42', model: 'Network'})": 1,
		"MATCH (:Sim:ModelInst)-[:Alias {name: 'signal_channel'}]->(:Sim:Channel {name: 'signal'})":    1,
		"MATCH (:Sim:Selector)-[:Identifies]->(:Sim:Channel)":                                          2,
	}},
	{"network.yaml", map[string]int64{
		"MATCH (:File {name: 'network.yaml'})-[:Contains]->(:Sim:Network {name: 'stub'})": 1,
		"MATCH (:Sim:Network)-[:Has]->(:Sim:Messages)":                                    2,
		"MATCH (:Sim:Messages)-[:Has]->(:Sim:Signals)":                                    4,
		"MATCH (:Sim:Messages)-[:Has]->(:Sim:Functions)-[:Has]->(:Sim:Encode)":            2,
		"MATCH (:Sim:Messages)-[:Has]->(:Sim:Functions)-[:Has]->(:Sim:Decode)":            1,
	}},
	{"runnable.yaml", map[string]int64{
		"MATCH (:File {name: 'runnable.yaml'})-[:Contains]->(:Sim:Runnable {name: 'target'})": 1,
		"MATCH (:Sim:Runnable)-[:Has]->(:Sim:Tasks)":                                          6,
	}},
	{"parameterset.yaml", map[string]int64{
		"MATCH (:File {name: 'parameterset.yaml'})-[:Contains]->(:Sim:ParameterSet {name: 'softecu_config_parameters'})": 1,
		"MATCH (:Sim:ParameterSet)-[:Has]->(:Sim:Parameters)":                                                            2,
	}},
	{"propagator.yaml", map[string]int64{
		"MATCH (:File {name: 'propagator.yaml'})-[:Contains]->(:Sim:Propagator {name: 'propagator'})":           1,
		"MATCH (:Sim:Propagator)-[:Has]->(:Sim:Signals {signal: 'foo', target: 'bar'})-[:Has]->(:Sim:Encoding)": 1,
		"MATCH (:Sim:Encoding)-[:Has]->(:Sim:Linear {min: 0.0, max: 10.0, factor: 0.5, offset: 2.0})":           1,
		"MATCH (:Sim:Encoding)-[:Has]->(:Sim:Mapping {name: 'map'})-[:Has]->(:Sim:Range)":                       1,
		"MATCH (:Sim:Propagator)-[:Has]->(:Sim:Options)":                                                        1,
	}},
	{"simulation.yaml", map[string]int64{
		"MATCH (:File {name: 'simulation.yaml'})-[:Contains]->(:Ast:Simulation {name: 'project'})": 1,
	}},
	{"manifest.yaml", map[string]int64{
		"MATCH (:File {name: 'manifest.yaml'})-[:Contains]->(:Sim:Manifest {name: 'fruity'})":   1,
		"MATCH (:Sim:Manifest)-[:Has]->(:Sim:Repo)":                                             1,
		"MATCH (:Sim:Manifest)-[:Has]->(:Sim:Tool)":                                             1,
		"MATCH (:Sim:Manifest)-[:Has]->(:Sim:Documentation)":                                    1,
		"MATCH (:Sim:Manifest)-[:Has]->(:Sim:Simulation {name: 'string'})-[:Has]->(:Sim:Files)": 1,
		"MATCH (:Sim:Simulation)-[:Has]->(:Sim:Models)-[:Has]->(:Sim:Channels)":                 1,
		"MATCH (:Sim:Manifest)-[:Has]->(:Sim:Model)-[:Has]->(:Sim:Channels)":                    1,
	}},
}

func TestImportYAML(t *testing.T) {
	for _, tt := range importTests {
		t.Run(tt.file, func(t *testing.T) {
			ctx, session := graphtest.Connect(t)

			handler, data, err := GetHandler(filepath.Join(yamlDir, tt.file))
			require.NoError(t, err)
			require.NoError(t, handler.Import(ctx, tt.file, data))

			for query, want := range tt.counts {
				got := graphtest.Count(t, ctx, session, query+" RETURN count(*) AS n", nil)
				assert.Equal(t, want, got, query)
			}
		})
	}
}

func TestImportIsIdempotent(t *testing.T) {
	ctx, session := graphtest.Connect(t)
	nodes := func() int64 {
		return graphtest.Count(t, ctx, session, "MATCH (n) RETURN count(n) AS n", nil)
	}
	rels := func() int64 {
		return graphtest.Count(t, ctx, session, "MATCH ()-[r]->() RETURN count(r) AS n", nil)
	}

	handler, data, err := GetHandler(filepath.Join(yamlDir, "stack.yaml"))
	require.NoError(t, err)
	require.NoError(t, handler.Import(ctx, "stack.yaml", data))
	firstNodes, firstRels := nodes(), rels()
	require.Greater(t, firstNodes, int64(1))

	require.NoError(t, handler.Import(ctx, "stack.yaml", data))
	assert.Equal(t, firstNodes, nodes())
	assert.Equal(t, firstRels, rels())
}

func TestImportMultiDocument(t *testing.T) {
	ctx, session := graphtest.Connect(t)
	path := writeFile(t, "multi.yaml", `---
kind: Runnable
metadata:
  name: first
spec:
  tasks:
    - function: task_a
      schedule: 1
---
kind: ParameterSet
metadata:
  name: second
spec:
  parameters:
    - parameter: p
      value: 1
`)

	handler, data, err := GetHandler(path)
	require.NoError(t, err)
	require.NoError(t, handler.Import(ctx, path, data))

	// Documents are linked from the file in document order.
	assert.Equal(t, int64(1), graphtest.Count(t, ctx, session,
		"MATCH (:File {name: 'multi.yaml'})-[:Contains {index: '1'}]->(:Sim:Runnable {name: 'first'}) RETURN count(*) AS n", nil))
	assert.Equal(t, int64(1), graphtest.Count(t, ctx, session,
		"MATCH (:File {name: 'multi.yaml'})-[:Contains {index: '2'}]->(:Sim:ParameterSet {name: 'second'}) RETURN count(*) AS n", nil))
}

func TestImportSkipsUnknownKind(t *testing.T) {
	ctx, session := graphtest.Connect(t)
	path := writeFile(t, "mixed.yaml", `---
kind: Bogus
metadata:
  name: nope
spec: {}
---
kind: Runnable
metadata:
  name: good
spec:
  tasks: []
`)

	handler, data, err := GetHandler(path)
	require.NoError(t, err)
	require.NoError(t, handler.Import(ctx, path, data))

	assert.Equal(t, int64(1), graphtest.Count(t, ctx, session, "MATCH (n:Sim:Runnable {name: 'good'}) RETURN count(*) AS n", nil))
	assert.Equal(t, int64(0), graphtest.Count(t, ctx, session, "MATCH (n {name: 'nope'}) RETURN count(*) AS n", nil))
}

func TestImportWithoutData(t *testing.T) {
	ctx, session := graphtest.Connect(t)
	handler := fileHandlers["text/yaml"][0]

	assert.Error(t, handler.Import(ctx, "none.yaml", nil))
	assert.Error(t, handler.Import(ctx, "none.yaml", "not a kind list"))
	assert.Equal(t, int64(0), graphtest.Count(t, ctx, session, "MATCH (n) RETURN count(n) AS n", nil))
}

func TestImportFailureRollsBackDocument(t *testing.T) {
	ctx, session := graphtest.Connect(t)
	path := writeFile(t, "partial.yaml", `---
kind: SignalGroup
metadata:
  name: bad
  annotations:
    graph:
      nodes:
        - label: Extra
          properties:
            name: x
spec:
  signals:
    - signal: s1
---
kind: Runnable
metadata:
  name: good
spec:
  tasks:
    - function: task_a
      schedule: 1
`)

	handler, data, err := GetHandler(path)
	require.NoError(t, err)
	err = handler.Import(ctx, path, data)

	// The failing document is rolled back, the other one is imported.
	require.Error(t, err)
	assert.Contains(t, err.Error(), "SignalGroup document 1")
	assert.Equal(t, int64(0), graphtest.Count(t, ctx, session, "MATCH (n:Sim:SignalGroup) RETURN count(n) AS n", nil))
	assert.Equal(t, int64(0), graphtest.Count(t, ctx, session, "MATCH (n:Sim:Signal) RETURN count(n) AS n", nil))
	assert.Equal(t, int64(1), graphtest.Count(t, ctx, session, "MATCH (:Sim:Runnable {name: 'good'})-[:Has]->(:Sim:Tasks) RETURN count(*) AS n", nil))
}

func TestGetHandlerErrors(t *testing.T) {
	tests := []struct {
		name    string
		path    func(t *testing.T) string
		message string
	}{
		{"missing file", func(t *testing.T) string { return filepath.Join(t.TempDir(), "absent.yaml") }, "unknown file type"},
		{"plain text", func(t *testing.T) string { return writeFile(t, "notes.txt", "just some text\n") }, "unsupported file"},
		{"binary", func(t *testing.T) string { return writeFile(t, "data.bin", "\x00\x01\x02\x03\xff") }, "unsupported file"},
		{"empty yaml", func(t *testing.T) string { return writeFile(t, "empty.yaml", "") }, "unsupported file"},
		{"only unknown kinds", func(t *testing.T) string {
			return writeFile(t, "unknown.yaml", "kind: Bogus\nmetadata:\n  name: x\nspec: {}\n")
		}, "unsupported file"},
		{"invalid yaml", func(t *testing.T) string { return writeFile(t, "broken.yaml", "kind: [unclosed\n  : :\n") }, "unsupported file"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := tt.path(t)
			done := make(chan struct{})
			var (
				handler Handler
				err     error
			)
			go func() {
				defer close(done)
				handler, _, err = GetHandler(path)
			}()
			select {
			case <-done:
			case <-time.After(10 * time.Second):
				t.Fatal("GetHandler did not return")
			}
			assert.Nil(t, handler)
			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.message)
		})
	}
}

func writeFile(t *testing.T, name, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), name)
	require.NoError(t, os.WriteFile(path, []byte(content), 0o644))
	return path
}

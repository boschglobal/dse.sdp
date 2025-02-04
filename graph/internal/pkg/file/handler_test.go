package file

import (
	"reflect"
	"testing"

	"github.com/boschglobal/dse.sdp/graph/internal/pkg/file/kind"
	"github.com/stretchr/testify/assert"
)

func TestHandleFile(t *testing.T) {
	tests := []struct {
		message   string
		path      string
		handler   Handler
		kind      string
		spec      any
		assertion assert.ComparisonAssertionFunc
	}{
		{
			"Kind::SignalGroup",
			"kind/testdata/yaml/signalgroup.yaml",
			&kind.YamlKindHandler{},
			"SignalGroup",
			&kind.SignalGroupSpec{},
			assert.Exactly,
		},
		{
			"Kind::Model",
			"kind/testdata/yaml/model.yaml",
			&kind.YamlKindHandler{},
			"Model",
			&kind.ModelSpec{},
			assert.Exactly,
		},
		{
			"Kind::Stack",
			"kind/testdata/yaml/stack.yaml",
			&kind.YamlKindHandler{},
			"Stack",
			&kind.StackSpec{},
			assert.Exactly,
		},
		{
			"Kind::Network",
			"kind/testdata/yaml/network.yaml",
			&kind.YamlKindHandler{},
			"Network",
			&kind.NetworkSpec{},
			assert.Exactly,
		},
		{
			"Kind::Runnable",
			"kind/testdata/yaml/runnable.yaml",
			&kind.YamlKindHandler{},
			"Runnable",
			&kind.RunnableSpec{},
			assert.Exactly,
		},
		{
			"Kind::Manifest",
			"kind/testdata/yaml/manifest.yaml",
			&kind.YamlKindHandler{},
			"Manifest",
			&kind.ManifestSpec{},
			assert.Exactly,
		},
		{
			"Kind::Propagator",
			"kind/testdata/yaml/propagator.yaml",
			&kind.YamlKindHandler{},
			"Propagator",
			&kind.PropagatorSpec{},
			assert.Exactly,
		},
		{
			"Kind::Simulation",
			"kind/testdata/yaml/simulation.yaml",
			&kind.YamlKindHandler{},
			"Simulation",
			&kind.SimulationSpec{},
			assert.Exactly,
		},
	}
	for _, test := range tests {
		t.Log(test.message)
		handler, data, err := GetHandler((test.path))
		assert.NotNil(t, handler)
		assert.NotNil(t, data)
		assert.NoError(t, err)
		test.assertion(t, test.handler, handler)

		if reflect.TypeOf(data).Kind() != reflect.Slice {
			data = []any{data}
		}
		t.Logf("%+v", data)
		if len(test.kind) > 0 {
			assert.Equal(t, test.kind, data.([]kind.KindDoc)[0].Kind)
		}
		if test.spec != nil {
			assert.IsType(t, test.spec, data.([]kind.KindDoc)[0].Spec)
		}
		//assert.False(t, true)

	}

}

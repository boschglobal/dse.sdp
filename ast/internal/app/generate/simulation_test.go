// Copyright 2025 Robert Bosch GmbH
//
// SPDX-License-Identifier: Apache-2.0

package generate

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func generateSimulation(t *testing.T, input string) string {
	data, err := os.ReadFile(input)
	assert.NoError(t, err)
	// Generate command resolves relative input paths under out/, so stage testdata there.
	stagedInput := filepath.Join("out", input)
	err = os.MkdirAll(filepath.Dir(stagedInput), 0755)
	assert.NoError(t, err)
	err = os.WriteFile(stagedInput, data, 0644)
	assert.NoError(t, err)

	// Keep output relative because the command prepends out/ internally.
	outFolder := filepath.Join("tmp", t.Name())
	err = os.RemoveAll(filepath.Join("out", outFolder))
	assert.NoError(t, err)
	err = os.MkdirAll(filepath.Join("out", outFolder), 0755)
	assert.NoError(t, err)
	simulationPath := filepath.Join("out", outFolder, "simulation.yaml")
	cmd := NewGenerateCommand("test_generate_simulation")
	args := []string{"-simulation", "-input", input, "-output", outFolder}

	err = cmd.Parse(args)
	assert.NoError(t, err)
	err = cmd.Run()
	assert.NoError(t, err)

	return simulationPath
}

func TestGenerateSimulation(t *testing.T) {
	simulationPath := generateSimulation(t, "testdata/ast__openloop.yaml")
	assert.FileExists(t, simulationPath)
	f, _ := os.ReadFile(simulationPath)
	t.Logf("\n%s\n", f)

	YamlContains(t, f, "$.kind", "Stack")
	YamlContains(t, f, "$.metadata.name", "default")

	YamlContains(t, f, "$.spec.connection.transport.redis.timeout", "60")
	YamlContains(t, f, "$.spec.connection.transport.redis.uri", "redis://localhost:6379")

	YamlContains(t, f, "$.spec.models[0].name", "simbus")
	YamlContains(t, f, "$.spec.models[0].uid", "0")
	YamlContains(t, f, "$.spec.models[0].model.name", "simbus")

	// The generated simulation YAML does not guarantee channel order.
	// Build a lookup by channel name to avoid index-based test failures.
	var doc map[string]interface{}
	require.NoError(t, yaml.Unmarshal(f, &doc))

	models := doc["spec"].(map[string]interface{})["models"].([]interface{})
	channels := models[0].(map[string]interface{})["channels"].([]interface{})

	counts := make(map[string]string)
	for _, c := range channels {
		ch := c.(map[string]interface{})
		counts[ch["name"].(string)] = fmt.Sprintf("%v", ch["expectedModelCount"])
	}

	assert.Equal(t, "2", counts["physical"])
	assert.Equal(t, "1", counts["network"])

	YamlContains(t, f, "$.spec.models[1].name", "input")
	YamlContains(t, f, "$.spec.models[1].uid", "1")
	YamlContains(t, f, "$.spec.models[1].model.name", "dse.modelc.csv")
	YamlContains(t, f, "$.spec.models[1].runtime.env.CSV_FILE", "model/input/data/input.csv")
	YamlContains(t, f, "$.spec.models[1].runtime.paths[0]", "model/input/data")
	YamlContains(t, f, "$.spec.models[1].channels[0].name", "physical")
	YamlContains(t, f, "$.spec.models[1].channels[0].alias", "signal_channel")
	YamlContains(t, f, "$.spec.models[1].channels[0].selectors.channel", "signal_vector")
	YamlContains(t, f, "$.spec.models[1].channels[0].selectors.model", "input")

	YamlContains(t, f, "$.spec.models[2].name", "linear")
	YamlContains(t, f, "$.spec.models[2].uid", "2")
	YamlContains(t, f, "$.spec.models[2].model.name", "linear")
	YamlContains(t, f, "$.spec.models[2].runtime.paths[0]", "model/linear/data")
	YamlContains(t, f, "$.spec.models[2].channels[0].name", "physical")
	YamlContains(t, f, "$.spec.models[2].channels[0].alias", "signal_channel")
	YamlContains(t, f, "$.spec.models[2].channels[0].selectors.channel", "signal_vector")
	YamlContains(t, f, "$.spec.models[2].channels[0].selectors.model", "linear")
	YamlContains(t, f, "$.spec.models[2].channels[1].name", "network")
	YamlContains(t, f, "$.spec.models[2].channels[1].alias", "network_channel")
	YamlContains(t, f, "$.spec.models[2].channels[1].selectors.channel", "network_vector")
	YamlContains(t, f, "$.spec.models[2].channels[1].selectors.model", "linear")

}

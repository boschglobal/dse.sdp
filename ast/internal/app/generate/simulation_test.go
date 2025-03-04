// Copyright 2025 Robert Bosch GmbH
//
// SPDX-License-Identifier: Apache-2.0

package generate

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func generateSimulation(t *testing.T, input string) string {
	var outFolder = t.TempDir()
	var simulationPath = filepath.Join(outFolder, "simulation.yaml")
	cmd := NewGenerateCommand("test_generate_simulation")
	args := []string{"-simulation", "-input", input, "-output", outFolder}

	err := cmd.Parse(args)
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
	YamlContains(t, f, "$.spec.models[0].channels[0].name", "physical")
	YamlContains(t, f, "$.spec.models[0].channels[0].expectedModelCount", "2")
	YamlContains(t, f, "$.spec.models[0].channels[1].name", "network")
	YamlContains(t, f, "$.spec.models[0].channels[1].expectedModelCount", "1")

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

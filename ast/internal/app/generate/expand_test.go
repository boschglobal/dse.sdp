// Copyright 2025 Robert Bosch GmbH
//
// SPDX-License-Identifier: Apache-2.0

package generate

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExpandTxtar(t *testing.T) {
	tmpDir := t.TempDir()
	txtarContent := `-- foo.txt --
Created
-- bar/foo.txt --
Created in folder bar
-- simulation.dse --
Not created, would overwrite DSE Script
-- ../foo.txt --
Not created, outside the project folder
-- /tmp/foo.txt --
Not created, outside the project folder
-- /foo.txt --
Not created, outside the project folder
`
	archivePath := filepath.Join(tmpDir, "simulation_txtar.dse")
	err := os.WriteFile(archivePath, []byte(txtarContent), 0644)
	require.NoError(t, err)

	err = ExpandTxtar(archivePath, tmpDir, true)
	require.NoError(t, err)

	cases := []struct {
		relPath     string
		want        string
		shouldExist bool
	}{
		{"foo.txt", "Created\n", true},
		{"bar/foo.txt", "Created in folder bar\n", true},
		{"simulation.dse", "", false},
		{"../foo.txt", "", false},
		{"/tmp/foo.txt", "", false},
	}
	for _, tc := range cases {
		path := filepath.Join(tmpDir, tc.relPath)
		data, err := os.ReadFile(path)
		if tc.shouldExist {
			assert.NoError(t, err, "file %s should exist", tc.relPath)
			assert.Equal(t, tc.want, string(data), "file %s content", tc.relPath)
		} else {
			assert.True(t, os.IsNotExist(err), "file %s should NOT exist", tc.relPath)
		}
	}

	preexisting := filepath.Join(tmpDir, "foo.txt")
	err = os.WriteFile(preexisting, []byte("PREEXISTING\n"), 0644)
	assert.NoError(t, err)
	_ = os.Remove(filepath.Join(tmpDir, "bar", "foo.txt"))

	err = ExpandTxtar(archivePath, tmpDir, false)
	assert.NoError(t, err)

	data, err := os.ReadFile(preexisting)
	assert.NoError(t, err)
	assert.Equal(t, "PREEXISTING\n", string(data))

	data, err = os.ReadFile(filepath.Join(tmpDir, "bar", "foo.txt"))
	assert.NoError(t, err)
	assert.Equal(t, "Created in folder bar\n", string(data))
}

func TestExpandTxtar_outputDirIsSeparateFromArchiveDir(t *testing.T) {
	archiveDir := t.TempDir()
	outputDir := t.TempDir()

	archivePath := filepath.Join(archiveDir, "sim.dse")
	err := os.WriteFile(archivePath, []byte("-- result.txt --\nhello\n"), 0644)
	require.NoError(t, err)

	err = ExpandTxtar(archivePath, outputDir, true)
	require.NoError(t, err)

	assert.FileExists(t, filepath.Join(outputDir, "result.txt"))
	assert.NoFileExists(t, filepath.Join(archiveDir, "result.txt"))
}

func TestExpandTxtar_simulationScenario(t *testing.T) {
	workDir := t.TempDir()
	simOutputDir := t.TempDir()

	dseScriptContent := `simulation
channel signal

uses
dse.sdp https://github.com/boschglobal/dse.sdp v1.0.0

model input Simple
    channel signal signal_vector
    file input.csv input.csv

-- input.csv --
Timestamp;input;factor;offset
0.0000;1.0;2.0;3.0
0.0005;-1.1;2.1;3.1
0.0010;1.2;-2.2;3.2
-- data/config.yaml --
key: value
nested:
  depth: 1
-- the.dse --
Not created (self-destruction prevention)
-- ../escape.txt --
Not created (path traversal)
-- /absolute.txt --
Not created (absolute path)
`
	dseScriptPath := filepath.Join(workDir, "the.dse")
	err := os.WriteFile(dseScriptPath, []byte(dseScriptContent), 0644)
	require.NoError(t, err)

	err = ExpandTxtar(dseScriptPath, simOutputDir, true)
	require.NoError(t, err)

	inputCSV := filepath.Join(simOutputDir, "input.csv")
	require.FileExists(t, inputCSV)
	data, err := os.ReadFile(inputCSV)
	require.NoError(t, err)
	assert.Equal(t,
		"Timestamp;input;factor;offset\n0.0000;1.0;2.0;3.0\n0.0005;-1.1;2.1;3.1\n0.0010;1.2;-2.2;3.2\n",
		string(data))

	subdirConfig := filepath.Join(simOutputDir, "data", "config.yaml")
	require.FileExists(t, subdirConfig)
	data, err = os.ReadFile(subdirConfig)
	require.NoError(t, err)
	assert.Equal(t, "key: value\nnested:\n  depth: 1\n", string(data))

	assert.NoFileExists(t, filepath.Join(simOutputDir, "the.dse"))
	assert.NoFileExists(t, filepath.Join(simOutputDir, "..", "escape.txt"))
	assert.NoFileExists(t, filepath.Join(simOutputDir, "absolute.txt"))

	err = os.WriteFile(inputCSV, []byte("SENTINEL\n"), 0644)
	require.NoError(t, err)
	err = os.Remove(subdirConfig)
	require.NoError(t, err)

	err = ExpandTxtar(dseScriptPath, simOutputDir, false)
	require.NoError(t, err)

	data, err = os.ReadFile(inputCSV)
	assert.NoError(t, err)
	assert.Equal(t, "SENTINEL\n", string(data))

	data, err = os.ReadFile(subdirConfig)
	assert.NoError(t, err)
	assert.Equal(t, "key: value\nnested:\n  depth: 1\n", string(data))

	err = ExpandTxtar(dseScriptPath, simOutputDir, true)
	require.NoError(t, err)

	data, err = os.ReadFile(inputCSV)
	require.NoError(t, err)
	assert.Equal(t,
		"Timestamp;input;factor;offset\n0.0000;1.0;2.0;3.0\n0.0005;-1.1;2.1;3.1\n0.0010;1.2;-2.2;3.2\n",
		string(data))
}

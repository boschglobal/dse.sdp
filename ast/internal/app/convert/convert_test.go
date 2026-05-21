// Copyright 2025 Robert Bosch GmbH
//
// SPDX-License-Identifier: Apache-2.0

package convert

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/boschglobal/dse.sdp/ast/internal/app/generate"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConvert_OriginalDseScriptLabel(t *testing.T) {
	dir := t.TempDir()
	jsonFile := filepath.Join(dir, "sim.json")
	dseFile := filepath.Join(dir, "sim.dse")
	outFile := filepath.Join(dir, "ast.yaml")

	err := os.WriteFile(jsonFile, []byte(`{}`), 0644)
	assert.NoError(t, err)
	err = os.WriteFile(dseFile, []byte("simulation\nchannel signal\n"), 0644)
	assert.NoError(t, err)

	cmd := NewConvertCommand("test_convert")
	err = cmd.Parse([]string{"-input", jsonFile, "-output", outFile})
	assert.NoError(t, err)
	err = cmd.Run()
	assert.NoError(t, err)

	data, err := os.ReadFile(outFile)
	assert.NoError(t, err)
	assert.True(t, strings.Contains(string(data), dseFile))
}

func TestConvert_OriginalDseScriptLabel_AbsentWhenDseFileMissing(t *testing.T) {
	dir := t.TempDir()
	jsonFile := filepath.Join(dir, "sim.json")
	outFile := filepath.Join(dir, "ast.yaml")

	err := os.WriteFile(jsonFile, []byte(`{}`), 0644)
	assert.NoError(t, err)

	cmd := NewConvertCommand("test_convert")
	err = cmd.Parse([]string{"-input", jsonFile, "-output", outFile})
	assert.NoError(t, err)
	err = cmd.Run()
	assert.NoError(t, err)

	data, err := os.ReadFile(outFile)
	assert.NoError(t, err)
	assert.False(t, strings.Contains(string(data), "original_dse_script"))
}

func TestExpandTxtar(t *testing.T) {
	tmpDir := t.TempDir()
	dseScript := filepath.Join(tmpDir, "simulation.dse")
	dseContent := `simulation
channel signal

-- foo.txt --
Created
-- bar/foo.txt --
Created in folder bar
-- simulation.dse --
Not created, would overwrite DSE Script (this file)
-- ../foo.txt --
Not created, outside the project folder
-- /tmp/foo.txt --
Not created, outside the project folder
-- /foo.txt --
Not created, outside the project folder
`
	err := os.WriteFile(dseScript, []byte(dseContent), 0644)
	require.NoError(t, err)

	err = generate.ExpandTxtar(dseScript, tmpDir, true)
	require.NoError(t, err)

	data, err := os.ReadFile(filepath.Join(tmpDir, "foo.txt"))
	require.NoError(t, err)
	assert.Equal(t, "Created\n", string(data))

	data, err = os.ReadFile(filepath.Join(tmpDir, "bar", "foo.txt"))
	require.NoError(t, err)
	assert.Equal(t, "Created in folder bar\n", string(data))

	orig, err := os.ReadFile(dseScript)
	require.NoError(t, err)
	assert.Equal(t, dseContent, string(orig))

	assert.NoFileExists(t, filepath.Join(tmpDir, "..", "foo.txt"))
	assert.NoFileExists(t, "/tmp/foo.txt")
	assert.NoFileExists(t, "/foo.txt")

	err = os.WriteFile(filepath.Join(tmpDir, "foo.txt"), []byte("ORIGINAL\n"), 0644)
	require.NoError(t, err)
	err = generate.ExpandTxtar(dseScript, tmpDir, false)
	require.NoError(t, err)
	data, err = os.ReadFile(filepath.Join(tmpDir, "foo.txt"))
	require.NoError(t, err)
	assert.Equal(t, "ORIGINAL\n", string(data))

	err = generate.ExpandTxtar(dseScript, tmpDir, true)
	require.NoError(t, err)
	data, err = os.ReadFile(filepath.Join(tmpDir, "foo.txt"))
	require.NoError(t, err)
	assert.Equal(t, "Created\n", string(data))
}

// Copyright 2025 Robert Bosch GmbH
//
// SPDX-License-Identifier: Apache-2.0

package generate

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/goccy/go-yaml"
	"github.com/stretchr/testify/assert"
)

func YamlContains(t *testing.T, file []byte, ypath string, text string) {
	path, err := yaml.PathString(ypath)
	assert.NoError(t, err, "Path no good: %s", ypath)

	var value string
	err = path.Read(strings.NewReader(string(file)), &value)
	assert.NoError(t, err, "Path not found in YAML: %s", ypath)
	assert.Equal(t, text, value, "Value did not match")
}

func TestLoadInputFile_none(t *testing.T) {
	c := GenerateCommand{}
	err := c.loadAst("")
	assert.Error(t, err)
}

func TestLoadInputFile_ast(t *testing.T) {
	c := GenerateCommand{}
	err := c.loadAst("testdata/ast.yaml")
	assert.NoError(t, err)
}

func TestGenerateTaskfile(t *testing.T) {
	var astFile = "testdata/ast.yaml"
	var outFolder = t.TempDir()
	var taskfileName = filepath.Join(outFolder, "Taskfile.yml")
	cmd := NewGenerateCommand("test_generate_taskfile")
	args := []string{"-taskfile", "-input", astFile, "-output", outFolder}

	// Run the command.
	err := cmd.Parse(args)
	assert.NoError(t, err)
	err = cmd.Run()
	assert.NoError(t, err)

	// Check the generated file.
	assert.DirExists(t, outFolder)
	assert.FileExists(t, taskfileName)
}

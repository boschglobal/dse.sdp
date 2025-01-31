// Copyright 2025 Robert Bosch GmbH
//
// SPDX-License-Identifier: Apache-2.0

package generate

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

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

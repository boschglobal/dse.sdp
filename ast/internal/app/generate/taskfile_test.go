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

func generateTaskfile(t *testing.T, input string) string {
	var outFolder = t.TempDir()
	var taskfileName = filepath.Join(outFolder, "Taskfile.yml")
	cmd := NewGenerateCommand("test_generate_taskfile")
	args := []string{"-taskfile", "-input", input, "-output", outFolder}

	err := cmd.Parse(args)
	assert.NoError(t, err)
	err = cmd.Run()
	assert.NoError(t, err)

	return taskfileName
}

func TestGenerateTaskfile_global_vars(t *testing.T) {
	taskfileName := generateTaskfile(t, "testdata/ast__global_vars.yaml")
	assert.FileExists(t, taskfileName)
	f, _ := os.ReadFile(taskfileName)
	t.Logf("\n%s\n", f)

	YamlContains(t, f, "$.version", "3")

	YamlContains(t, f, "$.vars.PLATFORM_ARCH", "linux-amd86")
	YamlContains(t, f, "$.vars.OUTDIR", "out")
	YamlContains(t, f, "$.vars.SIMDIR", "sim")
}

func TestGenerateTaskfile_includes(t *testing.T) {
	taskfileName := generateTaskfile(t, "testdata/ast__includes.yaml")
	require.FileExists(t, taskfileName)
	f, _ := os.ReadFile(taskfileName)
	t.Logf("\n%s\n", f)

	YamlContains(t, f, "$.version", "3")

	YamlContains(t, f, "$.includes.'dse.modelc-v2.1.15'.taskfile", "https://raw.githubusercontent.com/boschglobal/dse.modelc/refs/tags/v2.1.15/Taskfile.yml")
	YamlContains(t, f, "$.includes.'dse.modelc-v2.1.15'.dir", "{{.OUTDIR}}/{{.SIMDIR}}")
	YamlContains(t, f, "$.includes.'dse.modelc-v2.1.15'.vars.IMAGE_TAG", "2.1.15")
	YamlContains(t, f, "$.includes.'dse.modelc-v2.1.15'.vars.SIM", "{{.SIMDIR}}")
	YamlContains(t, f, "$.includes.'dse.modelc-v2.1.15'.vars.ENTRYWORKDIR", "{{.PWD}}/{{.OUTDIR}}")

}

func TestGenerateTaskfile_build_simulation(t *testing.T) {
	taskfileName := generateTaskfile(t, "testdata/ast.yaml")
	require.FileExists(t, taskfileName)
	f, _ := os.ReadFile(taskfileName)
	t.Logf("\n%s\n", f)

	YamlContains(t, f, "$.version", "3")

	YamlContains(t, f, "$.tasks.default.cmds[0].task", "build")

	YamlContains(t, f, "$.tasks.build.dir", "{{.OUTDIR}}")
	YamlContains(t, f, "$.tasks.build.label", "build")
	YamlContains(t, f, "$.tasks.build.deps[0].task", "build-models")
	YamlContains(t, f, "$.tasks.build.cmds[0]", "mkdir -p {{.SIMDIR}}/data")
	YamlContains(t, f, "$.tasks.build.cmds[1]", "cp {{.PWD}}/simulation.yaml {{.SIMDIR}}/data/simulation.yaml")
	YamlContains(t, f, "$.tasks.build.sources[0]", "{{.PWD}}/simulation.yaml")
	YamlContains(t, f, "$.tasks.build.generates[0]", "{{.SIMDIR}}/data/simulation.yaml")

}

func TestGenerateTaskfile_common_elements(t *testing.T) {
	taskfileName := generateTaskfile(t, "testdata/ast.yaml")
	require.FileExists(t, taskfileName)
	f, _ := os.ReadFile(taskfileName)
	t.Logf("\n%s\n", f)

	YamlContains(t, f, "$.version", "3")

	YamlContains(t, f, "$.tasks.unzip-file.dir", "{{.OUTDIR}}")
	YamlContains(t, f, "$.tasks.unzip-file.run", "when_changed")
	YamlContains(t, f, "$.tasks.unzip-file.label", "dse:unzip-file:{{.ZIPFILE}}-{{.FILEPATH}}")
	YamlContains(t, f, "$.tasks.unzip-file.vars.ZIP", "{{.ZIP}}")
	YamlContains(t, f, "$.tasks.unzip-file.vars.ZIPFILE", "{{.ZIPFILE}}")
	YamlContains(t, f, "$.tasks.unzip-file.vars.FILE", "{{.FILE}}")
	YamlContains(t, f, "$.tasks.unzip-file.cmds[3]", "mv -n $(dirname {{.FILE}})/$(basename {{.ZIPFILE}}) {{.FILE}}")
	YamlContains(t, f, "$.tasks.unzip-file.sources[0]", "{{.ZIP}}")
	YamlContains(t, f, "$.tasks.unzip-file.generates[0]", "{{.FILE}}")

	YamlContains(t, f, "$.tasks.unzip-dir.dir", "{{.OUTDIR}}")
	YamlContains(t, f, "$.tasks.unzip-dir.run", "when_changed")
	YamlContains(t, f, "$.tasks.unzip-dir.label", "dse:unzip-dir:{{.ZIPFILE}}-{{.DIR}}")
	YamlContains(t, f, "$.tasks.unzip-dir.vars.ZIP", "{{.ZIP}}")
	YamlContains(t, f, "$.tasks.unzip-dir.vars.ZIPDIR", "$(basename {{.ZIP}} {{ext .ZIP}})/{{.ZIPDIR}}")
	YamlContains(t, f, "$.tasks.unzip-dir.vars.DIR", "{{.DIR}}")
	YamlContains(t, f, "$.tasks.unzip-dir.cmds[2]", "unzip -o {{.ZIP}} {{.ZIPDIR}}/* -d {{.DIR}}")
	YamlContains(t, f, "$.tasks.unzip-dir.cmds[4]", "rm -rf {{.DIR}}/$(basename {{.ZIP}} {{ext .ZIP}})")
	YamlContains(t, f, "$.tasks.unzip-dir.sources[0]", "{{.ZIP}}")
	YamlContains(t, f, "$.tasks.unzip-dir.generates[0]", "{{.DIR}}/**")

	YamlContains(t, f, "$.tasks.unzip-extract-fmu.dir", "{{.OUTDIR}}")
	// FIXME add rest of this

	YamlContains(t, f, "$.tasks.download-file.dir", "{{.OUTDIR}}")
	YamlContains(t, f, "$.tasks.download-file.run", "when_changed")
	YamlContains(t, f, "$.tasks.download-file.label", "dse:download-file:{{.URL}}-{{.FILE}}")
	YamlContains(t, f, "$.tasks.download-file.vars.URL", "{{.URL}}")
	YamlContains(t, f, "$.tasks.download-file.vars.FILE", "{{.FILE}}")
	YamlContains(t, f, "$.tasks.download-file.vars.AUTH", "{{if all .USER .TOKEN}}-u {{.USER}}:{{.TOKEN}}{{else}}{{end}}")
	YamlContains(t, f, "$.tasks.download-file.cmds[2]", "curl --retry 5 {{.AUTH}} -fL {{.URL}} -o {{.FILE}}")
	YamlContains(t, f, "$.tasks.download-file.generates[0]", "{{.FILE}}")
	YamlContains(t, f, "$.tasks.download-file.status[0]", "test -f {{.FILE}}")

	YamlContains(t, f, "$.tasks.clean.cmds[0]", "find ./out -mindepth 1 -maxdepth 1 ! -name downloads -exec rm -rf {} +")

	YamlContains(t, f, "$.tasks.cleanall.cmds[0]", "rm -rf ./out")
}

func TestGenerateTaskfile_model_modelc(t *testing.T) {
	taskfileName := generateTaskfile(t, "testdata/ast__model_modelc.yaml")
	assert.FileExists(t, taskfileName)
	f, _ := os.ReadFile(taskfileName)
	t.Logf("\n%s\n", f)

	YamlContains(t, f, "$.tasks.build-models.label", "build-models")
	YamlContains(t, f, "$.tasks.build-models.deps[0].task", "model-input")

	YamlContains(t, f, "$.tasks.model-input.dir", "{{.OUTDIR}}")
	YamlContains(t, f, "$.tasks.model-input.label", "sim:model:input")

	YamlContains(t, f, "$.tasks.model-input.vars.REPO", "https://github.com/boschglobal/dse.modelc")
	YamlContains(t, f, "$.tasks.model-input.vars.TAG", "2.1.15")
	YamlContains(t, f, "$.tasks.model-input.vars.MODEL", "input")
	YamlContains(t, f, "$.tasks.model-input.vars.PATH", "model/input")
	YamlContains(t, f, "$.tasks.model-input.vars.PACKAGE_URL", "{{.REPO}}/releases/download/v{{.TAG}}/ModelC-{{.TAG}}-{{.PLATFORM_ARCH}}.zip")
	YamlContains(t, f, "$.tasks.model-input.vars.PACKAGE_PATH", "examples/csv")

	YamlContains(t, f, "$.tasks.model-input.deps[0].task", "download-file")
	YamlContains(t, f, "$.tasks.model-input.deps[0].vars.URL", "{{.PACKAGE_URL}}")
	YamlContains(t, f, "$.tasks.model-input.deps[0].vars.FILE", "downloads/{{base .PACKAGE_URL}}")
	YamlContains(t, f, "$.tasks.model-input.deps[1].task", "download-file")
	YamlContains(t, f, "$.tasks.model-input.deps[1].vars.URL", "http://some.server/fileshare/input.csv")
	YamlContains(t, f, "$.tasks.model-input.deps[1].vars.FILE", "downloads/input.csv")

	YamlContains(t, f, "$.tasks.model-input.cmds[0]", "echo \"SIM Model input -> {{.SIMDIR}}/{{.PATH}}\"")
	YamlContains(t, f, "$.tasks.model-input.cmds[1]", "mkdir -p '{{.SIMDIR}}/{{.PATH}}/data'")
	YamlContains(t, f, "$.tasks.model-input.cmds[2]", "cp {{.PWD}}/downloads/input.csv '{{.SIMDIR}}/{{.PATH}}/data/input.csv'")
	YamlContains(t, f, "$.tasks.model-input.cmds[3]", "cp {{.PWD}}/signalgroup.yaml '{{.SIMDIR}}/{{.PATH}}/data/signalgroup.yaml'")
	YamlContains(t, f, "$.tasks.model-input.cmds[4]", "cp '/volume/output.csv' {{.PWD}}/downloads/output.csv")
	YamlContains(t, f, "$.tasks.model-input.cmds[5]", "cp {{.PWD}}/downloads/output.csv '{{.SIMDIR}}/{{.PATH}}/trace/output.bmp'")

	YamlContains(t, f, "$.tasks.model-input.cmds[6].task", "unzip-dir")
	YamlContains(t, f, "$.tasks.model-input.cmds[6].vars.ZIP", "downloads/{{base .PACKAGE_URL}}")
	YamlContains(t, f, "$.tasks.model-input.cmds[6].vars.ZIPDIR", "{{.PACKAGE_PATH}}")
	YamlContains(t, f, "$.tasks.model-input.cmds[6].vars.DIR", "{{.SIMDIR}}/{{.PATH}}")

	YamlContains(t, f, "$.tasks.model-input.cmds[7]", "find {{.SIMDIR}}/{{.PATH}}/data -type f -name model.yaml -print0 | xargs -0 yq -i 'with(.spec.runtime.dynlib[]; .path |= sub(\".*/(.*$)\", \"{{.SIMDIR}}/{{.PATH}}/lib/${1}\"))'")
	YamlContains(t, f, "$.tasks.model-input.cmds[8]", "rm -rf '{{.SIMDIR}}/{{.PATH}}/examples'")

	YamlContains(t, f, "$.tasks.model-input.sources[0]", "{{.PWD}}/downloads/input.csv")
	YamlContains(t, f, "$.tasks.model-input.sources[1]", "{{.PWD}}/signalgroup.yaml")
	YamlContains(t, f, "$.tasks.model-input.sources[2]", "{{.PWD}}/downloads/output.csv")

	YamlContains(t, f, "$.tasks.model-input.generates[0]", "downloads/{{base .PACKAGE_URL}}")
	YamlContains(t, f, "$.tasks.model-input.generates[1]", "{{.SIMDIR}}/{{.PATH}}/data/input.csv")
	YamlContains(t, f, "$.tasks.model-input.generates[2]", "downloads/input.csv")
	YamlContains(t, f, "$.tasks.model-input.generates[3]", "{{.SIMDIR}}/{{.PATH}}/data/signalgroup.yaml")
	YamlContains(t, f, "$.tasks.model-input.generates[4]", "{{.SIMDIR}}/{{.PATH}}/trace/output.bmp")
	YamlContains(t, f, "$.tasks.model-input.generates[5]", "downloads/output.csv")
	YamlContains(t, f, "$.tasks.model-input.generates[6]", "{{.SIMDIR}}/{{.PATH}}/**")
}

func TestGenerateTaskfile_model_fmu(t *testing.T) {
	taskfileName := generateTaskfile(t, "testdata/ast__model_fmu.yaml")
	assert.FileExists(t, taskfileName)
	f, _ := os.ReadFile(taskfileName)
	t.Logf("\n%s\n", f)

	YamlContains(t, f, "$.tasks.build-models.label", "build-models")
	YamlContains(t, f, "$.tasks.build-models.deps[0].task", "model-linear")

	YamlContains(t, f, "$.tasks.model-linear.dir", "{{.OUTDIR}}")
	YamlContains(t, f, "$.tasks.model-linear.label", "sim:model:linear")

	YamlContains(t, f, "$.tasks.model-linear.vars.REPO", "https://github.com/boschglobal/dse.fmi")
	YamlContains(t, f, "$.tasks.model-linear.vars.TAG", "1.1.20")
	YamlContains(t, f, "$.tasks.model-linear.vars.MODEL", "linear")
	YamlContains(t, f, "$.tasks.model-linear.vars.PATH", "model/linear")
	YamlContains(t, f, "$.tasks.model-linear.vars.PACKAGE_URL", "{{.REPO}}/releases/download/v{{.TAG}}/Fmi-{{.TAG}}-{{.PLATFORM_ARCH}}.zip")
	YamlContains(t, f, "$.tasks.model-linear.vars.PACKAGE_PATH", "fmimcl")

	YamlContains(t, f, "$.tasks.model-linear.deps[0].task", "download-file")
	YamlContains(t, f, "$.tasks.model-linear.deps[0].vars.URL", "{{.PACKAGE_URL}}")
	YamlContains(t, f, "$.tasks.model-linear.deps[0].vars.FILE", "downloads/{{base .PACKAGE_URL}}")

	YamlContains(t, f, "$.tasks.model-linear.deps[1].task", "download-file")
	YamlContains(t, f, "$.tasks.model-linear.deps[1].vars.URL", "https://github.com/boschglobal/dse.fmi/releases/download/v1.1.20/Fmi-1.1.20-linux-amd64.zip")
	YamlContains(t, f, "$.tasks.model-linear.deps[1].vars.FILE", "downloads/Fmi-1.1.20-linux-amd64.zip")

	YamlContains(t, f, "$.tasks.model-linear.cmds[0]", "echo \"SIM Model linear -> {{.SIMDIR}}/{{.PATH}}\"")
	YamlContains(t, f, "$.tasks.model-linear.cmds[1]", "mkdir -p '{{.SIMDIR}}/{{.PATH}}/data'")

	YamlContains(t, f, "$.tasks.model-linear.cmds[2].task", "unzip-dir")
	YamlContains(t, f, "$.tasks.model-linear.cmds[2].vars.ZIP", "downloads/{{base .PACKAGE_URL}}")
	YamlContains(t, f, "$.tasks.model-linear.cmds[2].vars.ZIPDIR", "{{.PACKAGE_PATH}}")
	YamlContains(t, f, "$.tasks.model-linear.cmds[2].vars.DIR", "{{.SIMDIR}}/{{.PATH}}")

	YamlContains(t, f, "$.tasks.model-linear.cmds[3]", "find {{.SIMDIR}}/{{.PATH}}/data -type f -name model.yaml -print0 | xargs -0 yq -i 'with(.spec.runtime.dynlib[]; .path |= sub(\".*/(.*$)\", \"{{.SIMDIR}}/{{.PATH}}/lib/${1}\"))'")

	YamlContains(t, f, "$.tasks.model-linear.cmds[4]", "rm -rf '{{.SIMDIR}}/{{.PATH}}/examples'")
	YamlContains(t, f, "$.tasks.model-linear.cmds[5]", "find '{{.SIMDIR}}/{{.PATH}}' -type f -name simulation.yaml -print0  | xargs -0 rm -f")
	YamlContains(t, f, "$.tasks.model-linear.cmds[6]", "find '{{.SIMDIR}}/{{.PATH}}' -type f -name simulation.yml -print0  | xargs -0 rm -f")

	YamlContains(t, f, "$.tasks.model-linear.cmds[7].task", "unzip-extract-fmu")
	YamlContains(t, f, "$.tasks.model-linear.cmds[7].vars.ZIP", "downloads/Fmi-1.1.20-linux-amd64.zip")
	YamlContains(t, f, "$.tasks.model-linear.cmds[7].vars.FMUFILE", "examples/fmu/linear/fmi2/linear.fmu")
	YamlContains(t, f, "$.tasks.model-linear.cmds[7].vars.FMUDIR", "{{.SIMDIR}}/{{.PATH}}/linear_fmu")

	YamlContains(t, f, "$.tasks.model-linear.cmds[8].task", "dse.fmi-v1.1.20:generate-fmimcl")
	YamlContains(t, f, "$.tasks.model-linear.cmds[8].vars.FMU_DIR", "{{.PATH}}/linear_fmu")
	YamlContains(t, f, "$.tasks.model-linear.cmds[8].vars.OUT_DIR", "{{.PATH}}/data")
	YamlContains(t, f, "$.tasks.model-linear.cmds[8].vars.MCL_PATH", "{{.PATH}}/lib/libfmimcl.so")

	YamlContains(t, f, "$.tasks.model-linear.generates[0]", "downloads/{{base .PACKAGE_URL}}")
	YamlContains(t, f, "$.tasks.model-linear.generates[1]", "{{.SIMDIR}}/{{.PATH}}/**")
	YamlContains(t, f, "$.tasks.model-linear.generates[2]", "downloads/Fmi-1.1.20-linux-amd64.zip")
	YamlContains(t, f, "$.tasks.model-linear.generates[3]", "{{.SIMDIR}}/{{.PATH}}/linear_fmu")

	YamlContains(t, f, "$.tasks.model-linear.generates[4]", "{{.SIMDIR}}/{{.PATH}}/data/model.yaml")
	YamlContains(t, f, "$.tasks.model-linear.generates[5]", "{{.SIMDIR}}/{{.PATH}}/data/signalgroup.yaml")
}

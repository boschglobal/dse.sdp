// Copyright 2025 Robert Bosch GmbH
//
// SPDX-License-Identifier: Apache-2.0

package generate

import (
	"fmt"
	"strings"

	"github.com/elliotchance/orderedmap/v2"

	"github.com/boschglobal/dse.clib/extra/go/command/util"
	"github.com/boschglobal/dse.schemas/code/go/dse/ast"
)

func buildBaseTasks() map[string]Task {
	baseTasks := map[string]Task{
		"unzip-file": {
			Dir:   util.StringPtr("{{.OUTDIR}}"),
			Run:   util.StringPtr("when_changed"),
			Label: util.StringPtr("dse:unzip-file:{{.ZIPFILE}}-{{.FILEPATH}}"),
			Vars: func() *OMap {
				om := OMap{orderedmap.NewOrderedMap[string, string]()}
				om.Set("ZIP", "{{.ZIP}}")
				om.Set("ZIPFILE", "{{.ZIPFILE}}")
				om.Set("FILE", "{{.FILE}}")
				return &om
			}(),
			Cmds: &[]Cmd{
				{Cmd: "echo \"UNZIP FILE {{.ZIP}}/{{.ZIPFILE}} -> {{.FILE}}\""},
				{Cmd: "mkdir -p $(dirname {{.FILE}})"},
				{Cmd: `
if unzip -l {{.ZIP}} $(basename {{.ZIP}} {{ext .ZIP}})/{{.ZIPFILE}} >/dev/null 2>&1; then
	unzip -o -j {{.ZIP}} $(basename {{.ZIP}} {{ext .ZIP}})/{{.ZIPFILE}} -d $(dirname {{.FILE}})
elif unzip -l {{.ZIP}} {{.ZIPFILE}} >/dev/null 2>&1; then
	unzip -o -j {{.ZIP}} {{.ZIPFILE}} -d $(dirname {{.FILE}})
else
	echo "Error: {{.FILE}} not found in {{.ZIP}}" >&2
	exit 1
fi`},
				{Cmd: "mv -n $(dirname {{.FILE}})/$(basename {{.ZIPFILE}}) {{.FILE}}"},
			},
			Sources:   &[]string{"{{.ZIP}}"},
			Generates: &[]string{"{{.FILE}}"},
		},
		"unzip-dir": {
			Dir:   util.StringPtr("{{.OUTDIR}}"),
			Run:   util.StringPtr("when_changed"),
			Label: util.StringPtr("dse:unzip-dir:{{.ZIP}}-{{.DIR}}"),
			Vars: func() *OMap {
				om := OMap{orderedmap.NewOrderedMap[string, string]()}
				om.Set("ZIP", "{{.ZIP}}")
				om.Set("ZIPDIR", "$(basename {{.ZIP}} {{ext .ZIP}}){{if .ZIPDIR}}/{{.ZIPDIR}}{{end}}")
				om.Set("DIR", "{{.DIR}}")
				return &om
			}(),
			Cmds: &[]Cmd{
				{Cmd: "echo \"UNZIP DIR {{.ZIP}}/{{.ZIPDIR}} -> {{.DIR}}\""},
				{Cmd: "mkdir -p {{.DIR}}"},
				{Cmd: "unzip -o {{.ZIP}} {{.ZIPDIR}}/* -d {{.DIR}}"},
				{Cmd: "rsync -a {{.DIR}}/{{.ZIPDIR}}/ {{.DIR}}/"},
				{Cmd: "rm -rf {{.DIR}}/$(basename {{.ZIP}} {{ext .ZIP}})"},
			},
			Sources:   &[]string{"{{.ZIP}}"},
			Generates: &[]string{"{{.DIR}}/**"},
		},
		"unzip-dir-nopath": {
			Dir:   util.StringPtr("{{.OUTDIR}}"),
			Run:   util.StringPtr("when_changed"),
			Label: util.StringPtr("dse:unzip-dir-nopath:{{.ZIP}}-{{.DIR}}"),
			Vars: func() *OMap {
				om := OMap{orderedmap.NewOrderedMap[string, string]()}
				om.Set("ZIP", "{{.ZIP}}")
				om.Set("ZIPDIR", "{{.ZIPDIR}}")
				om.Set("DIR", "{{.DIR}}")
				return &om
			}(),
			Cmds: &[]Cmd{
				{Cmd: "echo \"UNZIP DIR {{.ZIP}}/{{.ZIPDIR}} -> {{.DIR}}\""},
				{Cmd: "mkdir -p {{.DIR}}"},
				{Cmd: "unzip -o {{.ZIP}} {{.ZIPDIR}}/* -d {{.DIR}}"},
			},
			Sources:   &[]string{"{{.ZIP}}"},
			Generates: &[]string{"{{.DIR}}/{{.ZIPDIR}}/**"},
		},
		"unzip-rootdir": {
			Dir:   util.StringPtr("{{.OUTDIR}}"),
			Run:   util.StringPtr("when_changed"),
			Label: util.StringPtr("dse:unzip-rootdir:{{.ZIPFILE}}-{{.DIR}}"),
			Vars: func() *OMap {
				om := OMap{orderedmap.NewOrderedMap[string, string]()}
				om.Set("ZIP", "{{.ZIP}}")
				om.Set("DIR", "{{.DIR}}")
				return &om
			}(),
			Cmds: &[]Cmd{
				{Cmd: "echo \"UNZIP DIR {{.ZIP}} -> {{.DIR}}\""},
				{Cmd: "mkdir -p {{.DIR}}"},
				{Cmd: "unzip -o {{.ZIP}} '*' -d {{.DIR}}"},
			},
			Sources:   &[]string{"{{.ZIP}}"},
			Generates: &[]string{"{{.DIR}}/**"},
		},
		"unzip-extract-fmu": {
			Dir:   util.StringPtr("{{.OUTDIR}}"),
			Run:   util.StringPtr("when_changed"),
			Label: util.StringPtr("dse:unzip-extract-fmu:{{.ZIP}}-{{.FMUDIR}}"),
			Vars: func() *OMap {
				om := OMap{orderedmap.NewOrderedMap[string, string]()}
				om.Set("ZIP", "{{.ZIP}}")
				om.Set("FMUFILE", "{{.FMUFILE}}")
				om.Set("FMUDIR", "{{.FMUDIR}}")
				om.Set("FMUTMPFILE", "{{.FMUDIR}}.tmp")
				return &om
			}(),
			Cmds: &[]Cmd{
				{Cmd: "echo \"UNZIP FMU {{.ZIP}}/{{.FMUFILE}} -> {{.FMUDIR}}\""},
				{
					Task: "unzip-file",
					Vars: &map[string]string{
						"ZIP":     "{{.ZIP}}",
						"ZIPFILE": "{{.FMUFILE}}",
						"FILE":    "{{.FMUTMPFILE}}",
					},
				},
				{
					Task: "unzip-rootdir",
					Vars: &map[string]string{
						"ZIP": "{{.FMUTMPFILE}}",
						"DIR": "{{.FMUDIR}}",
					},
				},
				{Cmd: "rm -f {{.FMUTMPFILE}}"},
			},
			Sources:   &[]string{"{{.ZIP}}"},
			Generates: &[]string{"{{.FMUDIR}}/**"},
		},
		"download-file": {
			Dir:   util.StringPtr("{{.OUTDIR}}"),
			Run:   util.StringPtr("when_changed"),
			Label: util.StringPtr("dse:download-file:{{.URL}}-{{.FILE}}"),
			Vars: func() *OMap {
				om := OMap{orderedmap.NewOrderedMap[string, string]()}
				om.Set("URL", "{{.URL}}")
				om.Set("FILE", "{{.FILE}}")
				om.Set("AUTH", "{{if all .USER .TOKEN}}-u {{.USER}}:{{.TOKEN}}{{else}}{{end}}")
				return &om
			}(),
			Cmds: &[]Cmd{
				{Cmd: "echo \"CURL {{.URL}} -> {{.FILE}}\""},
				{Cmd: "mkdir -p $(dirname {{.FILE}})"},
				{Cmd: "curl --retry 5 {{.AUTH}} -fL {{.URL}} -o {{.FILE}}"},
			},
			Generates: &[]string{"{{.FILE}}"},
			Status:    &[]string{"test -f {{.FILE}}"},
		},
		"copy-file": {
			Dir:   util.StringPtr("{{.OUTDIR}}"),
			Run:   util.StringPtr("when_changed"),
			Label: util.StringPtr("dse:copy-file:{{.URL}}-{{.FILE}}"),
			Vars: func() *OMap {
				om := OMap{orderedmap.NewOrderedMap[string, string]()}
				om.Set("URL", "{{.URL}}")
				om.Set("FILE", "{{.FILE}}")
				return &om
			}(),
			Cmds: &[]Cmd{
				{Cmd: "echo \"COPY {{.URL}} -> {{.FILE}}\""},
				{Cmd: "mkdir -p $(dirname {{.FILE}})"},
				{Cmd: "cp {{.URL}} {{.FILE}}"},
			},
			Sources:   &[]string{"{{.URL}}"},
			Generates: &[]string{"{{.FILE}}"},
		},
		"download-file-github-asset": {
			Dir:   util.StringPtr("{{.OUTDIR}}"),
			Run:   util.StringPtr("when_changed"),
			Label: util.StringPtr("dse:download-file-github-asset:{{.URL}}-{{.FILE}}"),
			Vars: func() *OMap {
				om := OMap{orderedmap.NewOrderedMap[string, string]()}
				om.Set("URL", "{{.URL}}")
				om.Set("FILE", "{{.FILE}}")
				om.Set("TOKEN", "{{.TOKEN}}")
				om.Set("ASSET_NAME", "{{.ASSET_NAME}}")
				om.Set("TAG", "{{.TAG}}")
				om.Set("API_URL", "{{.API_URL}}")
				return &om
			}(),
			Cmds: &[]Cmd{
				{Cmd: "echo \"CURL {{.URL}} -> {{.FILE}}\""},
				{Cmd: "mkdir -p $(dirname {{.FILE}})"},
				{Cmd: "REL_JSON=$(curl \\\n" +
					"  -H \"Accept: application/vnd.github+json\" \\\n" +
					"  -H \"Authorization: Bearer {{.TOKEN}}\" \\\n" +
					"  {{.API_URL}}/releases/tags/{{.TAG}}); \\\n" +
					"ASSET_URL=$(echo $REL_JSON | jq -r '.assets[] | select(.name | contains(\"{{.ASSET_NAME}}\")) | .url'); \\\n" +
					"curl \\\n" +
					"  -H \"Accept: application/octet-stream\" \\\n" +
					"  -H \"Authorization: Bearer {{.TOKEN}}\" \\\n" +
					"  -fL $ASSET_URL \\\n" +
					"  -o {{.FILE}}\n"},
			},
			Generates: &[]string{"{{.FILE}}"},
			Status:    &[]string{"test -f {{.FILE}}"},
		},
		"clean": {
			Cmds: &[]Cmd{
				{Cmd: "find ./out -mindepth 1 -maxdepth 1 ! -name downloads -exec rm -rf {} +"},
			},
		},
		"cleanall": {
			Cmds: &[]Cmd{
				{Cmd: "rm -rf ./out"},
			},
		},
	}
	return baseTasks
}

func buildSimulationTasks(simSpec ast.SimulationSpec) map[string]Task {
	// Build the _sequential_ build commands (order is important).
	// Sim folder setup.
	buildCmds := []Cmd{
		{Task: "build-setup-sim"},
	}
	// Simer download/deploy.
	for _, stack := range simSpec.Stacks {
		if stack.Arch != nil && strings.HasPrefix(*stack.Arch, "windows-") {
			buildCmds = append(buildCmds, Cmd{
				Task: "deploy-simer-windows",
			})
			break
		}
	}
	// Stacks.
	for _, stack := range simSpec.Stacks {
		if stack.Name == "external" {
			continue
		}
		buildCmds = append(buildCmds, Cmd{
			Task: fmt.Sprintf("stack-%s", stack.Name), // Stack task is generated elsewhere.
		})
	}

	// Construct the simulation tasks.
	simulationTasks := map[string]Task{
		"default": {
			Cmds: &[]Cmd{
				{Task: "build"},
			},
		},
		"build": {
			Dir:   util.StringPtr("{{.OUTDIR}}"),
			Label: util.StringPtr("build"),
			Cmds:  &buildCmds,
		},
		"build-setup-sim": {
			Dir:   util.StringPtr("{{.OUTDIR}}"),
			Label: util.StringPtr("build-setup-sim"),
			Cmds: &[]Cmd{
				{Cmd: "mkdir -p {{.SIMDIR}}/data"},
				{Cmd: "cp {{.ENTRYDIR}}/simulation.yaml {{.SIMDIR}}/data/simulation.yaml"},
			},
			Sources:   &[]string{"{{.ENTRYDIR}}/simulation.yaml"},
			Generates: &[]string{"{{.SIMDIR}}/data/simulation.yaml"},
		},
	}
	for k, v := range generateSimerWindows(simSpec) {
		simulationTasks[k] = v
	}
	return simulationTasks
}

func generateSimerWindows(simSpec ast.SimulationSpec) map[string]Task {
	simerUses := locateSimerUses(simSpec)
	if simerUses == nil {
		return map[string]Task{}
	}
	simerTasks := map[string]Task{
		"deploy-simer-windows": {
			Dir:   util.StringPtr("{{.OUTDIR}}"),
			Label: util.StringPtr("dse:modelc:deploy-simer-windows"),
			Vars: func() *OMap {
				om := OMap{orderedmap.NewOrderedMap[string, string]()}
				om.Set("REPO", simerUses.Url)
				if simerUses.Version != nil {
					om.Set("TAG", cleanTag(*simerUses.Version))
				}
				om.Set("PATH", "{{.SIMDIR}}")
				om.Set("PACKAGE_URL", "{{.REPO}}/releases/download/v{{.TAG}}/Simer-{{.TAG}}-windows.zip")
				return &om
			}(),
			Deps: &[]Dep{
				{
					Task: "download-file",
					Vars: func() *OMap {
						om := OMap{orderedmap.NewOrderedMap[string, string]()}
						om.Set("URL", "{{.PACKAGE_URL}}")
						om.Set("FILE", "downloads/{{base .PACKAGE_URL}}")
						return &om
					}(),
				},
			},
			Cmds: &[]Cmd{
				{Cmd: "echo \"DOWNLOAD SIMER WINDOWS {{.PACKAGE_URL}} -> {{.PATH}}\""},
				{
					Task: "unzip-dir-nopath",
					Vars: &map[string]string{
						"DIR":    "{{.PATH}}",
						"ZIP":    "downloads/{{base .PACKAGE_URL}}",
						"ZIPDIR": "bin",
					},
				},
				{
					Task: "unzip-dir-nopath",
					Vars: &map[string]string{
						"DIR":    "{{.PATH}}",
						"ZIP":    "downloads/{{base .PACKAGE_URL}}",
						"ZIPDIR": "licenses",
					},
				},
			},
			Sources: &[]string{},
			Generates: &[]string{
				"downloads/{{base .PACKAGE_URL}}",
				"{{.PATH}}/bin/**",
				"{{.PATH}}/licenses/**",
			},
		},
	}

	return simerTasks
}

func locateSimerUses(simSpec ast.SimulationSpec) *ast.Uses {
	if simSpec.Uses == nil {
		return nil
	}
	usesName := "dse.modelc"
	for _, stack := range simSpec.Stacks {
		if stack.Arch != nil && strings.HasPrefix(*stack.Arch, "windows-") {
			val := getSimerUsesAnnotation(stack.Annotations)
			if val != "" {
				usesName = val
				break
			}
		}
	}
	for _, uses := range *simSpec.Uses {
		if uses.Name == usesName {
			return &uses
		}
	}
	return nil
}

func getSimerUsesAnnotation(annotations *ast.Annotations) string {
	if annotations != nil {
		if v, ok := (*annotations)["simer-uses-selector"]; ok {
			if s, ok := v.(string); ok {
				return s
			}
		}
	}
	return ""
}

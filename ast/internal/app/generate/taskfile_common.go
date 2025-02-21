// Copyright 2025 Robert Bosch GmbH
//
// SPDX-License-Identifier: Apache-2.0

package generate

import (
	"github.com/elliotchance/orderedmap/v2"
)

func buildBaseTasks() map[string]Task {
	baseTasks := map[string]Task{
		"unzip-file": {
			Dir:   stringPtr("{{.OUTDIR}}"),
			Run:   stringPtr("when_changed"),
			Label: stringPtr("dse:unzip-file:{{.ZIPFILE}}-{{.FILEPATH}}"),
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
				{Cmd: "unzip -o -j {{.ZIP}} $(basename {{.ZIP}} {{ext .ZIP}})/{{.ZIPFILE}} -d $(dirname {{.FILE}})"},
				{Cmd: "mv -n $(dirname {{.FILE}})/$(basename {{.ZIPFILE}}) {{.FILE}}"},
			},
			Sources:   &[]string{"{{.ZIP}}"},
			Generates: &[]string{"{{.FILE}}"},
		},
		"unzip-dir": {
			Dir:   stringPtr("{{.OUTDIR}}"),
			Run:   stringPtr("when_changed"),
			Label: stringPtr("dse:unzip-dir:{{.ZIPFILE}}-{{.DIR}}"),
			Vars: func() *OMap {
				om := OMap{orderedmap.NewOrderedMap[string, string]()}
				om.Set("ZIP", "{{.ZIP}}")
				om.Set("ZIPDIR", "{{if .ZIPDIR}}\"{{.ZIPDIR}}/*\"{{else}}{{end}}")
				om.Set("DIR", "{{.DIR}}")
				om.Set("JUNKDIR", "{{if .ZIPDIR}}-j{{else}}{{end}}")
				return &om
			}(),
			Cmds: &[]Cmd{
				{Cmd: "echo \"UNZIP DIR {{.ZIP}}/{{.ZIPDIR}} -> {{.DIR}}\""},
				{Cmd: "mkdir -p {{.DIR}}"},
				{Cmd: "unzip -o {{.JUNKDIR}} {{.ZIP}} {{.ZIPDIR}} -d {{.DIR}}"},
			},
			Sources:   &[]string{"{{.ZIP}}"},
			Generates: &[]string{"{{.DIR}}/**"},
		},
		"unzip-extract-fmu": {
			Dir:   stringPtr("{{.OUTDIR}}"),
			Run:   stringPtr("when_changed"),
			Label: stringPtr("dse:unzip-extract-fmu:{{.ZIP}}-{{.FMUDIR}}"),
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
					Task: "unzip-dir",
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
			Dir:   stringPtr("{{.OUTDIR}}"),
			Run:   stringPtr("when_changed"),
			Label: stringPtr("dse:download-file:{{.URL}}-{{.FILE}}"),
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

func buildSimulationTasks() map[string]Task {

	simulationTasks := map[string]Task{
		"default": {
			Cmds: &[]Cmd{
				{Task: "build"},
			},
		},
		"build": {
			Dir:   stringPtr("{{.OUTDIR}}"),
			Label: stringPtr("build"),
			Deps: &[]Dep{
				{Task: "build-models"},
			},
			Cmds: &[]Cmd{
				{Cmd: "mkdir -p {{.SIMDIR}}/data"},
				{Cmd: "cp {{.PWD}}/simulation.yaml {{.SIMDIR}}/data/simulation.yaml"},
			},
			Sources:   &[]string{"{{.PWD}}/simulation.yaml"},
			Generates: &[]string{"{{.SIMDIR}}/data/simulation.yaml"},
		},
	}
	return simulationTasks
}

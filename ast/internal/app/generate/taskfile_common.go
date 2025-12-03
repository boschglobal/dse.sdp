// Copyright 2025 Robert Bosch GmbH
//
// SPDX-License-Identifier: Apache-2.0

package generate

import (
	"github.com/elliotchance/orderedmap/v2"

	"github.com/boschglobal/dse.clib/extra/go/command/util"
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
				{Cmd: "unzip -o -j {{.ZIP}} $(basename {{.ZIP}} {{ext .ZIP}})/{{.ZIPFILE}} -d $(dirname {{.FILE}})"},
				{Cmd: "mv -n $(dirname {{.FILE}})/$(basename {{.ZIPFILE}}) {{.FILE}}"},
			},
			Sources:   &[]string{"{{.ZIP}}"},
			Generates: &[]string{"{{.FILE}}"},
		},
		"unzip-dir": {
			Dir:   util.StringPtr("{{.OUTDIR}}"),
			Run:   util.StringPtr("when_changed"),
			Label: util.StringPtr("dse:unzip-dir:{{.ZIPFILE}}-{{.DIR}}"),
			Vars: func() *OMap {
				om := OMap{orderedmap.NewOrderedMap[string, string]()}
				om.Set("ZIP", "{{.ZIP}}")
				om.Set("ZIPDIR", "$(basename {{.ZIP}} {{ext .ZIP}})/{{.ZIPDIR}}")
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

func buildSimulationTasks() map[string]Task {

	simulationTasks := map[string]Task{
		"default": {
			Cmds: &[]Cmd{
				{Task: "build"},
			},
		},
		"build": {
			Dir:   util.StringPtr("{{.OUTDIR}}"),
			Label: util.StringPtr("build"),
			Cmds: &[]Cmd{
				{Cmd: "mkdir -p {{.SIMDIR}}/data"},
				{Cmd: "cp {{.ENTRYDIR}}/simulation.yaml {{.SIMDIR}}/data/simulation.yaml"},
				{Task: "build-models"},
			},
			Sources:   &[]string{"{{.ENTRYDIR}}/simulation.yaml"},
			Generates: &[]string{"{{.SIMDIR}}/data/simulation.yaml"},
		},
	}
	return simulationTasks
}

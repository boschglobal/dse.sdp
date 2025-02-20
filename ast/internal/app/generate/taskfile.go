// Copyright 2025 Robert Bosch GmbH
//
// SPDX-License-Identifier: Apache-2.0

package generate

import (
	"flag"
	"fmt"
	"net/url"
	"path/filepath"
	"regexp"

	"github.boschdevcloud.com/fsil/fsil.go/command/util"
	"github.com/boschglobal/dse.schemas/code/go/dse/ast"
)

type Cmd struct {
	Cmd  string             `yaml:"cmd,omitempty"`
	Task string             `yaml:"task,omitempty"`
	Vars *map[string]string `yaml:"vars,omitempty"`
}

func (c Cmd) MarshalYAML() (interface{}, error) {
	if c.Cmd != "" {
		return c.Cmd, nil
	}
	if c.Task != "" {
		if c.Vars == nil {
			return map[string]any{"task": c.Task}, nil
		} else {
			return map[string]any{"task": c.Task, "vars": c.Vars}, nil
		}
	}
	return nil, nil
}

type Dep struct {
	Task string             `yaml:"task"`
	Vars *map[string]string `yaml:"vars,omitempty"`
}

type Requires struct {
	Vars *[]string `yaml:"vars,omitempty"`
}

type Task struct {
	Cmds      *[]Cmd             `yaml:"cmds,omitempty"`
	Desc      *string            `yaml:"desc,omitempty"`
	Dir       *string            `yaml:"dir,omitempty"`
	Label     *string            `yaml:"label,omitempty"`
	Requires  *Requires          `yaml:"requires,omitempty"`
	Run       *string            `yaml:"run,omitempty"`
	Vars      *map[string]string `yaml:"vars,omitempty"`
	Deps      *[]Dep             `yaml:"deps,omitempty"`
	Sources   *[]string          `yaml:"sources,omitempty"`
	Generates *[]string          `yaml:"generates,omitempty"`
	Status    *[]string          `yaml:"status,omitempty"`
}

type Include struct {
	Taskfile string             `yaml:"taskfile,omitempty"`
	Dir      string             `yaml:"dir,omitempty"`
	Vars     *map[string]string `yaml:"vars,omitempty"`
}

type Taskfile struct {
	Version  string              `yaml:"version"`
	Includes *map[string]Include `yaml:"includes,omitempty"`
	Vars     *map[string]string  `yaml:"vars,omitempty"`
	Tasks    *map[string]Task    `yaml:"tasks,omitempty"`
}

func stringPtr(s string) *string {
	return &s
}

func cleanTag(tag string) string {
	var clean = regexp.MustCompile(`[^0-9\.]+`)
	return clean.ReplaceAllString(tag, "")
}

func (c GenerateCommand) GenerateTaskfile() error {
	var taskfilePath = filepath.Join(c.outputPath, "Taskfile.yml")

	fmt.Fprintf(flag.CommandLine.Output(), "Writing taskfile: %s\n", taskfilePath)

	taskfile := Taskfile{
		Version: "3",
		Vars: &map[string]string{
			"PLATFORM_ARCH": func() string {
				if c.simulationAst.Arch != "" {
					return c.simulationAst.Arch
				} else {
					return "linux-amd64"
				}
			}(),
			"OUTDIR": "out",
			"SIMDIR": "sim",
		},
	}

	// FIXME MCL tasks
	//	MCL path
	//  The FMU itself
	//  Tasks -
	//		Fetch FMU : from uses ... each should be downloaded
	//		Workflows : from AST
	// Need to know its an FMU or at least MCL ... so that the correct deps task can be selected.
	//	TODO use the model name == dse.fmi.mcl as the selector.
	// Then the tasks; setup MCL, run Workflow (from AST)
	// Sources and generates ... more complex ... only partial on generates (the mcl).

	// FIXME MIMEtypes on channels.

	// FIXME Uses ... referenced by workflows (references: uses), part of model.
	//		need to resolve reference_uses and generate fetch/download with version tag in path.

	includes := make(map[string]Include)
	for k, v := range c.buildIncludes() {
		includes[k] = v
	}

	tasks := make(map[string]Task)
	for k, v := range buildSimulationTasks() {
		tasks[k] = v
	}
	modelTasks, err := c.buildModelTasks()
	if err != nil {
		return err
	}
	for k, v := range modelTasks {
		tasks[k] = v
	}
	for k, v := range buildBaseTasks() {
		tasks[k] = v
	}

	// FIXME need a task to correct paths in model.yaml files.

	taskfile.Tasks = &tasks
	taskfile.Includes = &includes
	if err := util.WriteYaml(&taskfile, taskfilePath, true); err != nil {
		return err
	}
	return nil
}

func (c GenerateCommand) buildIncludes() map[string]Include {
	includes := make(map[string]Include)
	simSpec := c.simulationAst

	if simSpec.Uses == nil {
		return includes
	}

	for _, uses := range *simSpec.Uses {
		if *&uses.Metadata == nil {
			continue
		}
		md := *uses.Metadata
		mdContainer, ok := md["container"]
		if !ok {
			continue
		}

		vars := map[string]string{
			"SIM":          "{{.SIMDIR}}",
			"ENTRYWORKDIR": "{{.PWD}}/{{.OUTDIR}}",
		}
		if imageKey, ok := mdContainer.(map[string]interface{})["image_var"]; ok {
			if imageVal, ok := mdContainer.(map[string]interface{})["repository"]; ok {
				vars[imageKey.(string)] = imageVal.(string)
			}
		}
		if tagKey, ok := mdContainer.(map[string]interface{})["tag_var"]; ok {
			vars[tagKey.(string)] = cleanTag(*uses.Version)
		}
		includes[uses.Name] = Include{
			Taskfile: func() string {
				u, _ := url.Parse(uses.Url)
				return fmt.Sprintf("https://raw.githubusercontent.com%s/refs/tags/%s/Taskfile.yml", u.Path, *uses.Version)
			}(),
			Dir:  "{{.OUTDIR}}/{{.SIMDIR}}",
			Vars: &vars,
		}
	}

	return includes
}

func genericModelTask(model ast.Model, modelUses ast.Uses) Task {
	deps := []Dep{
		{
			Task: "download-file",
			Vars: &map[string]string{
				"URL":  "{{.PACKAGE_URL}}",
				"FILE": "downloads/{{base .PACKAGE_URL}}",
			},
		},
	}
	cmds := []Cmd{
		{
			Cmd: fmt.Sprintf("echo \"SIM Model %s -> {{.SIMDIR}}/{{.PATH}}\"", model.Name),
		},
		{
			Cmd: "mkdir -p '{{.SIMDIR}}/{{.PATH}}/data'",
		},
	}
	sources := []string{}
	generates := []string{
		"downloads/{{base .PACKAGE_URL}}",
	}
	md := *model.Metadata

	modelTask := Task{
		Dir:   stringPtr("{{.OUTDIR}}"),
		Label: stringPtr(fmt.Sprintf("sim:model:%s", model.Name)),
		Vars: &map[string]string{
			"REPO":         modelUses.Url,
			"TAG":          cleanTag(*modelUses.Version),
			"MODEL":        model.Name,
			"PATH":         fmt.Sprintf("model/%s", model.Name),
			"PACKAGE_URL":  md["package"].(map[string]interface{})["download"].(string),
			"PACKAGE_PATH": md["models"].(map[string]interface{})[model.Model].(map[string]interface{})["path"].(string),
			// TODO need PLATFORM_ARCH if specified on Stack or Model
			// TODO need correction to files .. like model.yaml
		},
		Deps:      &deps,
		Cmds:      &cmds,
		Sources:   &sources,
		Generates: &generates,
	}
	return modelTask
}

func buildModel(model ast.Model, simSpec ast.SimulationSpec) (Task, error) {
	var modelUses *ast.Uses

	for _, uses := range *simSpec.Uses {
		if uses.Name == model.Uses {
			modelUses = &uses
			break
		}
	}
	if modelUses == nil {
		return Task{}, fmt.Errorf("model uses not found in simulation AST (name=%s)", "dse.modelc")
	}

	md := *model.Metadata
	modelTask := genericModelTask(model, *modelUses)

	// Parse: user files
	func(task *Task, model ast.Model) {
		if model.Files != nil {
			for _, file := range *model.Files {
				*task.Cmds = append(*task.Cmds, Cmd{
					Cmd: fmt.Sprintf("cp {{.PWD}}/%[1]s '{{.SIMDIR}}/{{.PATH}}/data/%[1]s'", file),
				})
				*task.Sources = append(*task.Sources, fmt.Sprintf("{{.PWD}}/%s", file))
				*task.Generates = append(*task.Generates, fmt.Sprintf("{{.SIMDIR}}/{{.PATH}}/data/%s", file))
			}
		}
	}(&modelTask, model)

	// Parse: modelc package/model files
	func(task *Task, model ast.Model) {
		modelFiles := md["models"].(map[string]interface{})[model.Model].(map[string]interface{})["files"].([]interface{})

		for _, file := range modelFiles {
			*task.Cmds = append(*task.Cmds, Cmd{
				Task: "unzip-file",
				Vars: &map[string]string{
					"ZIP":     "downloads/{{base .PACKAGE_URL}}",
					"ZIPFILE": fmt.Sprintf("{{.PACKAGE_PATH}}/%s", file.(string)),
					"FILE":    fmt.Sprintf("{{.SIMDIR}}/{{.PATH}}/%s", file.(string)),
				},
			})
			*task.Generates = append(*task.Generates, fmt.Sprintf("{{.SIMDIR}}/{{.PATH}}/%s", file.(string)))
		}
	}(&modelTask, model)

	// Parse: workflow uses items
	func(task *Task, model ast.Model) {
		if model.Workflows == nil {
			return
		}
		for _, workflow := range *model.Workflows {
			if workflow.Vars == nil {
				continue
			}
			for _, v := range *workflow.Vars {
				if v.Reference != nil && *v.Reference == "uses" {
					var varUses *ast.Uses
					for _, uses := range *simSpec.Uses {
						if uses.Name == v.Value {
							varUses = &uses
							break
						}
					}
					if varUses == nil {
						continue
					}
					// Download the Uses file.
					*task.Deps = append(*task.Deps, Dep{
						Task: "download-file",
						Vars: &map[string]string{
							"URL":  varUses.Url,
							"FILE": "downloads/{{base .URL}}",
						},
					})
					u, _ := url.Parse(varUses.Url)
					downloadFile := fmt.Sprintf("downloads/%s", filepath.Base(u.Path))
					*task.Generates = append(*task.Generates, downloadFile)
					// Extract the Uses path.
					if varUses.Path == nil {
						continue
					}
					if filepath.Ext(*varUses.Path) == ".fmu" {
						*task.Cmds = append(*task.Cmds, Cmd{
							Task: "unzip-extract-fmu",
							Vars: &map[string]string{
								"ZIP":     downloadFile,
								"FMUFILE": *varUses.Path,
								"FMUDIR":  fmt.Sprintf("{{.SIMDIR}}/{{.PATH}}/%s", varUses.Name),
							},
						})
						*task.Generates = append(*task.Generates, fmt.Sprintf("{{.SIMDIR}}/{{.PATH}}/%s", varUses.Name))
					}
				}
			}
		}
	}(&modelTask, model)

	// TODO Parse: workflow emit tasks
	func(task *Task, model ast.Model) {
		if model.Workflows == nil {
			return
		}
		for _, workflow := range *model.Workflows {
			var workflowUses *ast.Uses = modelUses
			if workflow.Uses != nil {
				for _, uses := range *simSpec.Uses {
					if uses.Name == *workflow.Uses {
						workflowUses = &uses
						break
					}
				}
			}
			vars := map[string]string{}
			if workflow.Vars == nil {
				continue
			}
			for _, v := range *workflow.Vars {
				if v.Reference != nil && *v.Reference == "uses" {
					name := fmt.Sprintf("%s_USES_VALUE", v.Name)
					vars[name] = v.Value
					vars[v.Name] = fmt.Sprintf("{{.SIMDIR}}/{{.PATH}}/{{.%s}}", name)
				} else {
					vars[v.Name] = v.Value
				}
			}
			*task.Cmds = append(*task.Cmds, Cmd{
				Task: fmt.Sprintf("%s-%s:%s", workflowUses.Name, *workflowUses.Version, workflow.Name),
				Vars: &vars,
			})
			workflowFiles := md["workflows"].(map[string]interface{})[workflow.Name].(map[string]interface{})["generates"].([]interface{})
			for _, file := range workflowFiles {
				*task.Generates = append(*task.Generates, fmt.Sprintf("{{.SIMDIR}}/{{.PATH}}/%s", file.(string)))
			}
		}
	}(&modelTask, model)

	return modelTask, nil
}

func (c GenerateCommand) buildModelTasks() (map[string]Task, error) {
	modelTaskNames := []string{}
	modelTasks := map[string]Task{}

	simSpec := c.simulationAst

	for _, stack := range simSpec.Stacks {
		for _, model := range stack.Models {
			modelName := fmt.Sprintf("model-%s", model.Name)
			modelTaskNames = append(modelTaskNames, modelName)
			mt, err := buildModel(model, simSpec)
			if err != nil {
				// FIXME what to do here ?
				fmt.Fprint(flag.CommandLine.Output(), fmt.Errorf("ERROR: build model task (%s): %w", model.Name, err))
				continue
				//return map[string]Task{}, err
			}
			modelTasks[modelName] = mt
		}
	}

	modelTasks["build-models"] = Task{
		Label: stringPtr("build-models"),
		Deps: func() *[]Dep {
			deps := []Dep{}
			for _, modelName := range modelTaskNames {
				deps = append(deps, Dep{Task: modelName})
			}
			return &deps
		}(),
	}

	return modelTasks, nil
}

func buildBaseTasks() map[string]Task {

	baseTasks := map[string]Task{
		"unzip-file": {
			Dir:   stringPtr("{{.OUTDIR}}"),
			Run:   stringPtr("when_changed"),
			Label: stringPtr("dse:unzip-file:{{.ZIPFILE}}-{{.FILEPATH}}"), // FIXME need to prefix the ZIPFILE with the name of the ZIP which is the root dir of the extracted content.
			Vars: &map[string]string{
				"ZIP":     "{{.ZIP}}",
				"ZIPFILE": "{{.ZIPFILE}}",
				"FILE":    "{{.FILE}}",
			},
			Cmds: &[]Cmd{
				{Cmd: "echo \"UNZIP FILE {{.ZIP}}/{{.ZIPFILE}} -> {{.FILE}}\""},
				{Cmd: "mkdir -p $(dirname {{.FILE}})"},
				{Cmd: "unzip -o -j {{.ZIP}} {{.ZIPFILE}} -d $(dirname {{.FILE}})"},
				{Cmd: "mv -n $(dirname {{.FILE}})/$(basename {{.ZIPFILE}}) {{.FILE}}"},
			},
			Sources:   &[]string{"{{.ZIP}}"},
			Generates: &[]string{"{{.FILE}}"},
		},
		"unzip-dir": {
			Dir:   stringPtr("{{.OUTDIR}}"),
			Run:   stringPtr("when_changed"),
			Label: stringPtr("dse:unzip-dir:{{.ZIPFILE}}-{{.DIR}}"),
			Vars: &map[string]string{
				"ZIP":     "{{.ZIP}}",
				"ZIPDIR":  "{{if .ZIPDIR}}\"{{.ZIPDIR}}/*\"{{else}}{{end}}",
				"DIR":     "{{.DIR}}",
				"JUNKDIR": "{{if .ZIPDIR}}-j{{else}}{{end}}",
			},
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
			Vars: &map[string]string{
				"ZIP":        "{{.ZIP}}",
				"FMUFILE":    "{{.FMUFILE}}",
				"FMUDIR":     "{{.FMUDIR}}",
				"FMUTMPFILE": "{{.FMUFILE}}.tmp",
			},
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
			},
			Sources:   &[]string{"{{.ZIP}}"},
			Generates: &[]string{"{{.FMUDIR}}/**"},
		},
		"download-file": {
			Dir:   stringPtr("{{.OUTDIR}}"),
			Run:   stringPtr("when_changed"),
			Label: stringPtr("dse:download-file:{{.URL}}-{{.FILE}}"),
			Vars: &map[string]string{
				"URL":  "{{.URL}}",
				"FILE": "{{.FILE}}",
				"AUTH": "{{if all .USER .TOKEN}}-u {{.USER}}:{{.TOKEN}}{{else}}{{end}}",
			},
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

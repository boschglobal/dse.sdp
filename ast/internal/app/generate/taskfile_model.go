// Copyright 2025 Robert Bosch GmbH
//
// SPDX-License-Identifier: Apache-2.0

package generate

import (
	"fmt"
	"net/url"
	"path/filepath"
	"strings"

	"github.com/elliotchance/orderedmap/v2"

	"github.boschdevcloud.com/fsil/fsil.go/command/util"
	"github.com/boschglobal/dse.schemas/code/go/dse/ast"
)

func (c GenerateCommand) buildIncludes() map[string]Include {
	includes := make(map[string]Include)
	simSpec := c.simulationAst

	if simSpec.Uses == nil {
		return includes
	}

	for _, uses := range *simSpec.Uses {
		if uses.Version == nil {
			continue
		}
		vars := map[string]string{
			"SIM":          "{{.SIMDIR}}",
			"ENTRYWORKDIR": "{{.PWD}}/{{.OUTDIR}}",
			"IMAGE_TAG":    cleanTag(*uses.Version),
		}
		includes[fmt.Sprintf("%s-%s", uses.Name, *uses.Version)] = Include{
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
			Vars: func() *OMap {
				om := OMap{orderedmap.NewOrderedMap[string, string]()}
				om.Set("URL", "{{.PACKAGE_URL}}")
				om.Set("FILE", "downloads/{{base .PACKAGE_URL}}")
				if modelUses.User != nil {
					if strings.HasPrefix(*modelUses.User, "$") {
						om.Set("USER", fmt.Sprintf("{{.%s}}", (*modelUses.User)[1:]))
					} else {
						om.Set("USER", *modelUses.User)
					}
				}
				if modelUses.Token != nil {
					if strings.HasPrefix(*modelUses.Token, "$") {
						om.Set("TOKEN", fmt.Sprintf("{{.%s}}", (*modelUses.Token)[1:]))
					} else {
						om.Set("TOKEN", *modelUses.Token)
					}
				}
				return &om
			}(),
		},
	}
	cmds := []Cmd{
		{
			Cmd: fmt.Sprintf("echo \"SIM Model %s -> {{.SIMDIR}}/{{.PATH}}\"", model.Name),
		},
		{
			Cmd: "mkdir -p {{.SIMDIR}}/{{.PATH}}/data",
		},
	}
	sources := []string{}
	generates := []string{
		"downloads/{{base .PACKAGE_URL}}",
	}
	md := map[string]interface{}{}
	if model.Metadata != nil {
		md = *model.Metadata
	}

	modelTask := Task{
		Dir:   util.StringPtr("{{.OUTDIR}}"),
		Label: util.StringPtr(fmt.Sprintf("sim:model:%s", model.Name)),
		Vars: func() *OMap {
			om := OMap{orderedmap.NewOrderedMap[string, string]()}
			if modelUses.Name != "" {
				om.Set("REPO", modelUses.Url)
				if modelUses.Version != nil {
					om.Set("TAG", cleanTag(*modelUses.Version))
				}
			}
			om.Set("MODEL", model.Name)
			om.Set("PATH", fmt.Sprintf("model/%s", model.Name))
			om.Set("PLATFORM_ARCH", *model.Arch)

			func() {
				defer func() {
					if r := recover(); r != nil {
					}
				}()
				om.Set("PACKAGE_URL", md["package"].(map[string]interface{})["download"].(string))
				om.Set("PACKAGE_PATH", md["models"].(map[string]interface{})[model.Model].(map[string]interface{})["path"].(string))
			}()

			return &om
		}(),
		Deps:      &deps,
		Cmds:      &cmds,
		Sources:   &sources,
		Generates: &generates,
	}
	return modelTask
}

func parseUrl(task *Task, uses *ast.Uses) string {
	u, _ := url.Parse(uses.Url)
	downloadFile := fmt.Sprintf("downloads/%s", filepath.Base(u.Path))
	if u.IsAbs() == true {
		*task.Deps = append(*task.Deps, Dep{
			Task: "download-file",
			Vars: func() *OMap {
				om := OMap{orderedmap.NewOrderedMap[string, string]()}
				om.Set("URL", uses.Url)
				om.Set("FILE", downloadFile)
				return &om
			}(),
		})
	} else {
		if filepath.IsAbs(uses.Url) {
			*task.Cmds = append(*task.Cmds, Cmd{
				Cmd: fmt.Sprintf("cp %s %s", uses.Url, downloadFile),
			})
		} else {
			*task.Cmds = append(*task.Cmds, Cmd{
				Cmd: fmt.Sprintf("cp {{.ENTRYDIR}}/%s %s", uses.Url, downloadFile),
			})
		}
	}
	*task.Generates = append(*task.Generates, downloadFile)
	return downloadFile
}

func buildModel(model ast.Model, simSpec ast.SimulationSpec) (Task, error) {
	var modelUses ast.Uses

	if len(model.Uses) > 0 {
		for _, uses := range *simSpec.Uses {
			if uses.Name == model.Uses {
				modelUses = uses
				break
			}
		}
		if modelUses.Name == "" {
			return Task{}, fmt.Errorf("model uses not found in simulation AST (name=%s)", model.Uses)
		}
	}

	md := map[string]interface{}{}
	if model.Metadata != nil {
		md = *model.Metadata
	}
	modelTask := genericModelTask(model, modelUses)

	// Parse: user files
	func(task *Task, model ast.Model) {
		if model.Files != nil {
			for _, f := range *model.Files {
				// Calculate the model relative path (i.e. generates).
				dir, file := filepath.Split(f.Name)
				if len(dir) == 0 {
					dir = "data/"
				}
				filePath := fmt.Sprintf("{{.SIMDIR}}/{{.PATH}}/%s%s", dir, file)
				*task.Generates = append(*task.Generates, fmt.Sprintf("%s", filePath))
				// Determine the source (i.e. sources).
				if f.Reference != nil && *f.Reference == "uses" {
					var fileUses *ast.Uses
					for _, uses := range *simSpec.Uses {
						if uses.Name == f.Value {
							fileUses = &uses
							break
						}
					}
					if fileUses == nil {
						continue
					}
					downloadFile := parseUrl(task, fileUses)
					*task.Cmds = append(*task.Cmds, Cmd{
						Cmd: fmt.Sprintf("cp %s %s", downloadFile, filePath),
					})
					*task.Sources = append(*task.Sources, fmt.Sprintf("%s", downloadFile))
				} else {
					if filepath.IsAbs(f.Value) {
						*task.Cmds = append(*task.Cmds, Cmd{
							Cmd: fmt.Sprintf("cp %s %s", f.Value, filePath),
						})
					} else {
						*task.Cmds = append(*task.Cmds, Cmd{
							Cmd: fmt.Sprintf("cp {{.ENTRYDIR}}/%s %s", f.Value, filePath),
						})
					}
					*task.Sources = append(*task.Sources, fmt.Sprintf("%s", f.Name))
				}
			}
		}
	}(&modelTask, model)

	// Parse: modelc package/model files
	func(task *Task, model ast.Model) {
		modelPath := ""
		func() {
			defer func() {
				if r := recover(); r != nil {
				}
			}()
			modelPath = md["models"].(map[string]interface{})[model.Model].(map[string]interface{})["path"].(string)
		}()
		if len(modelPath) != 0 {
			*task.Cmds = append(*task.Cmds, Cmd{
				Task: "unzip-dir",
				Vars: &map[string]string{
					"ZIP":    "downloads/{{base .PACKAGE_URL}}",
					"ZIPDIR": fmt.Sprintf("{{.PACKAGE_PATH}}"),
					"DIR":    fmt.Sprintf("{{.SIMDIR}}/{{.PATH}}"),
				},
			})
			*task.Cmds = append(*task.Cmds, Cmd{
				Cmd: fmt.Sprintf("find {{.SIMDIR}}/{{.PATH}}/data -type f -name model.yaml -print0 | " +
					"xargs -r -0 yq -i " +
					"'with(.spec.runtime.dynlib[]; " +
					".path |= sub(\".*/(.*$)\", \"{{.PATH}}/lib/${1}\"))'"),
			})
			*task.Cmds = append(*task.Cmds, Cmd{
				Cmd: fmt.Sprintf("rm -rf {{.SIMDIR}}/{{.PATH}}/examples"),
			})
			*task.Cmds = append(*task.Cmds, Cmd{
				Cmd: fmt.Sprintf("find {{.SIMDIR}}/{{.PATH}} -type f -name simulation.yaml -print0  | xargs -r -0 rm -f"),
			})
			*task.Cmds = append(*task.Cmds, Cmd{
				Cmd: fmt.Sprintf("find {{.SIMDIR}}/{{.PATH}} -type f -name simulation.yml -print0  | xargs -r -0 rm -f"),
			})
			*task.Generates = append(*task.Generates, fmt.Sprintf("{{.SIMDIR}}/{{.PATH}}/**"))
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
					downloadFile := parseUrl(task, varUses)

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

	// Parse: workflow emit tasks
	func(task *Task, model ast.Model) {
		if model.Workflows == nil {
			return
		}
		for _, workflow := range *model.Workflows {
			var workflowUses ast.Uses = modelUses
			if workflow.Uses != nil {
				for _, uses := range *simSpec.Uses {
					if uses.Name == *workflow.Uses {
						workflowUses = uses
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
					vars[v.Name] = fmt.Sprintf("{{.PATH}}/%s", v.Value)
				} else {
					vars[v.Name] = v.Value
				}
			}
			var workflowTaskName string
			if workflowUses.Version == nil {
				workflowTaskName = fmt.Sprintf("%s:%s", workflowUses.Name, workflow.Name)
			} else {
				workflowTaskName = fmt.Sprintf("%s-%s:%s", workflowUses.Name, *workflowUses.Version, workflow.Name)
			}
			*task.Cmds = append(*task.Cmds, Cmd{
				Task: workflowTaskName,
				Vars: &vars,
			})
			workflowFiles := []interface{}{}
			func() {
				defer func() {
					if r := recover(); r != nil {
					}
				}()
				workflowFiles = md["tasks"].(map[string]interface{})[workflow.Name].(map[string]interface{})["generates"].([]interface{})
			}()
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
				return nil, fmt.Errorf("Error building model (name=%s): %w", modelName, err)
			}
			modelTasks[modelName] = mt
		}
	}

	modelTasks["build-models"] = Task{
		Label: util.StringPtr("build-models"),
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

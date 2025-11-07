// Copyright 2025 Robert Bosch GmbH
//
// SPDX-License-Identifier: Apache-2.0

package generate

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/elliotchance/orderedmap/v2"

	"github.com/boschglobal/dse.clib/extra/go/command/util"
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
		if uses.Path != nil {
			continue
		}
		vars := map[string]string{
			"SIM":          "{{.SIMDIR}}",
			"ENTRYWORKDIR": "{{.PWD}}/{{.OUTDIR}}",
			"IMAGE_TAG":    cleanTag(*uses.Version),
		}
		if uses.User != nil {
			vars["DOCKER_USER"] = *uses.User
		}
		if uses.Token != nil {
			vars["DOCKER_TOKEN"] = *uses.Token
		}

		u, _ := func() (*url.URL, error) {
			_u := uses.Url
			_u = strings.ReplaceAll(_u, `{`, `%7B`)
			_u = strings.ReplaceAll(_u, `}`, `%7D`)
			return url.Parse(_u)
		}()
		if strings.HasPrefix(u.Host, "github.") == false {
			continue
		}

		includes[fmt.Sprintf("%s-%s", uses.Name, *uses.Version)] = Include{
			Taskfile: func() string {
				var finalUrl *url.URL
				pathParts := strings.Split(u.Path, string(os.PathSeparator))
				switch u.Host {
				case "github.com":
					u.Host = "raw.githubusercontent.com"
					finalUrl, _ = u.Parse(fmt.Sprintf("/%s/%s/refs/tags/%s/Taskfile.yml", pathParts[1], pathParts[2], *uses.Version))
				case "github.boschdevcloud.com":
					u.Host = "raw.github.boschdevcloud.com"
					finalUrl, _ = u.Parse(fmt.Sprintf("/%s/%s/%s/Taskfile.yml", pathParts[1], pathParts[2], *uses.Version))
				default:
					panic("unsupported includes URL")
				}
				return func() string {
					_u := finalUrl.String()
					_u = strings.ReplaceAll(_u, `%7B`, `{`)
					_u = strings.ReplaceAll(_u, `%7D`, `}`)
					return _u
				}()
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
					userValue := *modelUses.User
					if strings.HasPrefix(userValue, "$") {
						om.Set("USER", fmt.Sprintf("{{.%s}}", userValue[1:]))
					} else {
						om.Set("USER", userValue)
					}
				}
				if modelUses.Token != nil {
					tokenValue := *modelUses.Token
					if strings.HasPrefix(tokenValue, "$") {
						om.Set("TOKEN", fmt.Sprintf("{{.%s}}", tokenValue[1:]))
					} else {
						om.Set("TOKEN", tokenValue)
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
			if model.Arch != nil {
				om.Set("PLATFORM_ARCH", *model.Arch)
			}

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

func parseUrl(task *Task, uses *ast.Uses, modelName string) string {
	u, _ := url.Parse(uses.Url)
	downloadFile := fmt.Sprintf("downloads/models/{{.MODEL}}/%s", filepath.Base(u.Path))

	if u.IsAbs() == true {
		if strings.HasPrefix(u.Host, "github.boschdevcloud.") {
			// Rewrite the URL and fetch the GitHub Asset (using PAT authentication).
			*task.Deps = append(*task.Deps, Dep{
				Task: "download-file-github-asset",
				Vars: func() *OMap {
					om := OMap{orderedmap.NewOrderedMap[string, string]()}
					om.Set("URL", uses.Url)
					om.Set("FILE", downloadFile)
					if uses.Token != nil {
						om.Set("TOKEN", *uses.Token)
					}
					pathParts := strings.Split(u.Path, string(os.PathSeparator))
					om.Set("ASSET_NAME", pathParts[len(pathParts)-1])
					om.Set("TAG", pathParts[len(pathParts)-2])
					om.Set("API_URL", func() string {
						url, _ := u.Parse(fmt.Sprintf("/api/v3/repos/%s/%s", pathParts[1], pathParts[2]))
						return url.String()
					}())
					return &om
				}(),
			})
		} else {
			*task.Deps = append(*task.Deps, Dep{
				Task: "download-file",
				Vars: func() *OMap {
					om := OMap{orderedmap.NewOrderedMap[string, string]()}
					om.Set("URL", uses.Url)
					om.Set("FILE", downloadFile)
					if uses.User != nil {
						userValue := *uses.User
						if strings.HasPrefix(userValue, "$") {
							om.Set("USER", fmt.Sprintf("{{.%s}}", userValue[1:]))
						} else {
							om.Set("USER", userValue)
						}
					}
					if uses.Token != nil {
						tokenValue := *uses.Token
						if strings.HasPrefix(tokenValue, "$") {
							om.Set("TOKEN", fmt.Sprintf("{{.%s}}", tokenValue[1:]))
						} else {
							om.Set("TOKEN", tokenValue)
						}
					}
					return &om
				}(),
			})
		}
	} else {
		if filepath.IsAbs(uses.Url) {
			*task.Cmds = append(*task.Cmds,
				Cmd{Cmd: fmt.Sprintf("mkdir -p $(dirname %s)", downloadFile)},
				Cmd{Cmd: fmt.Sprintf("cp %s %s", uses.Url, downloadFile)},
			)
		} else {
			*task.Cmds = append(*task.Cmds,
				Cmd{Cmd: fmt.Sprintf("mkdir -p $(dirname %s)", downloadFile)},
				Cmd{Cmd: fmt.Sprintf("cp {{.ENTRYDIR}}/%s %s", uses.Url, downloadFile)},
			)
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
			return Task{}, fmt.Errorf("Model uses not found in simulation AST (name=%s)", model.Uses)
		}
	}
	usesDownloadFilePaths := map[string]string{}

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
					downloadFile := parseUrl(task, fileUses, model.Name)
					usesDownloadFilePaths[fileUses.Name] = downloadFile
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
		isModel := func() bool {
			defer func() {
				if r := recover(); r != nil {
				}
			}()
			if v := md["models"].(map[string]interface{})[model.Model].(map[string]interface{})["path"]; v != nil {
				return true
			} else {
				return false
			}
		}
		if isModel() == false {
			return
		}
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
					downloadFile := parseUrl(task, varUses, model.Name)
					usesDownloadFilePaths[varUses.Name] = downloadFile

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
						usesDownloadFilePaths[varUses.Name] = fmt.Sprintf("{{.PATH}}/%s", varUses.Name)
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
			var workflowUses *ast.Uses = nil
			// Search for the requested/specified 'uses'.
			if workflow.Uses != nil {
				for _, uses := range *simSpec.Uses {
					if uses.Name == *workflow.Uses {
						workflowUses = &uses
						break
					}
				}
			}
			// Otherwise search the 'uses' space for the workflow.
			// TODO should we only rely on explicit 'uses'?
			if workflowUses == nil {
				for _, uses := range *simSpec.Uses {
					if uses.Metadata == nil {
						continue
					}
					usesMd := *uses.Metadata
					if _, ok := usesMd["models"]; !ok {
						continue
					}
					models := usesMd["models"].(map[string]interface{})
					for _, model := range models {
						for _, w := range model.(map[string]interface{})["workflows"].([]interface{}) {
							if w.(string) == workflow.Name {
								workflowUses = &uses
							}
						}
					}
				}
			}
			// And lastly use the modelUses.
			if workflowUses == nil {
				workflowUses = &modelUses
			}

			vars := map[string]string{"MODEL": "{{.MODEL}}"}
			if workflow.Vars == nil {
				continue
			}
			for _, v := range *workflow.Vars {
				if v.Reference != nil && *v.Reference == "uses" {
					vars[v.Name] = usesDownloadFilePaths[v.Value]
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

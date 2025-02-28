// Copyright 2025 Robert Bosch GmbH
//
// SPDX-License-Identifier: Apache-2.0

package generate

import (
	"bytes"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"regexp"

	"github.com/elliotchance/orderedmap/v2"
	//	"github.com/elliotchance/orderedmap/v2"

	"github.boschdevcloud.com/fsil/fsil.go/command/util"
)

type OMap struct {
	*orderedmap.OrderedMap[string, string]
}

func (vm OMap) MarshalYAML() (interface{}, error) {
	s := ""
	for _, k := range vm.Keys() {
		v, _ := vm.Get(k)
		s = fmt.Sprintf("%s\n%s: '%s'", s, k, v)
		// TODO make the value more robust {{}} => '{{}}' and other cases.
	}
	return s, nil
}

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
	Task string `yaml:"task"`
	//Vars *map[string]string `yaml:"vars,omitempty"`
	//Vars *orderedmap.OrderedMap[string, string]
	Vars *OMap `yaml:"vars,omitempty"`
}

type Requires struct {
	Vars *[]string `yaml:"vars,omitempty"`
}

type Task struct {
	Cmds      *[]Cmd    `yaml:"cmds,omitempty"`
	Desc      *string   `yaml:"desc,omitempty"`
	Dir       *string   `yaml:"dir,omitempty"`
	Label     *string   `yaml:"label,omitempty"`
	Requires  *Requires `yaml:"requires,omitempty"`
	Run       *string   `yaml:"run,omitempty"`
	Vars      *OMap     `yaml:"vars,omitempty"`
	Deps      *[]Dep    `yaml:"deps,omitempty"`
	Sources   *[]string `yaml:"sources,omitempty"`
	Generates *[]string `yaml:"generates,omitempty"`
	Status    *[]string `yaml:"status,omitempty"`
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

func cleanTag(tag string) string {
	var clean = regexp.MustCompile(`[^0-9\.]+`)
	return clean.ReplaceAllString(tag, "")
}

func (c GenerateCommand) GenerateTaskfile() error {
	var taskfilePath = filepath.Join(c.outputPath, "Taskfile.yml")

	fmt.Fprintf(flag.CommandLine.Output(), "Writing taskfile: %s\n", taskfilePath)

	// Setup the basic Taskfile structure.
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
			"ENTRYDIR": "{{if .SIM}}{{.ENTRYWORKDIR}}/{{.SIM}}{{else}}{{.PWD}}{{end}}",
			"OUTDIR":   "out",
			"SIMDIR":   "sim",
		},
	}
	tasks := make(map[string]Task)
	for k, v := range buildSimulationTasks() {
		tasks[k] = v
	}
	for k, v := range buildBaseTasks() {
		tasks[k] = v
	}

	// Build the Model Tasks and associated Includes.
	includes := make(map[string]Include)
	for k, v := range c.buildIncludes() {
		includes[k] = v
	}
	modelTasks, err := c.buildModelTasks()
	if err != nil {
		return err
	}
	for k, v := range modelTasks {
		tasks[k] = v
	}

	// TODO MIMEtypes on channels.

	// Finalise the Taskfile.
	taskfile.Tasks = &tasks
	taskfile.Includes = &includes
	if err := util.WriteYaml(&taskfile, taskfilePath, false); err != nil {
		return err
	}
	// Correct sorted Vars in the generated YAML.
	data, err := os.ReadFile(taskfilePath)
	if err != nil {
		return err
	}
	os.WriteFile(taskfilePath, bytes.ReplaceAll(data, []byte("vars: |2-"), []byte("vars:")), 0644)

	return nil
}

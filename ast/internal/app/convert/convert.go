// Copyright 2025 Robert Bosch GmbH
//
// SPDX-License-Identifier: Apache-2.0

package convert

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log/slog"
	"os"

	"github.com/tidwall/gjson"

	"github.boschdevcloud.com/fsil/fsil.go/command"
	"github.boschdevcloud.com/fsil/fsil.go/command/log"
	"github.boschdevcloud.com/fsil/fsil.go/command/util"

	"github.com/boschglobal/dse.schemas/code/go/dse/ast"
)

type ConvertCommand struct {
	command.Command

	inputFile  string
	outputFile string
	logLevel   int

	dslAst []byte
}

func NewConvertCommand(name string) *ConvertCommand {
	c := &ConvertCommand{
		Command: command.Command{
			Name:    name,
			FlagSet: flag.NewFlagSet(name, flag.ExitOnError),
		},
	}
	c.FlagSet().StringVar(&c.inputFile, "input", "", "path to DSL generated AST file")
	c.FlagSet().StringVar(&c.outputFile, "output", "", "path to write generated AST file")
	c.FlagSet().IntVar(&c.logLevel, "log", 4, "Loglevel")
	return c
}

func (c ConvertCommand) Name() string {
	return c.Command.Name
}

func (c ConvertCommand) FlagSet() *flag.FlagSet {
	return c.Command.FlagSet
}

func (c *ConvertCommand) Parse(args []string) error {
	return c.FlagSet().Parse(args)
}

func (c *ConvertCommand) Run() error {
	slog.SetDefault(log.NewLogger(c.logLevel))

	fmt.Fprintf(flag.CommandLine.Output(), "Reading file: %s\n", c.inputFile)
	fmt.Fprintf(flag.CommandLine.Output(), "Writing file: %s\n", c.outputFile)

	c.loadDslAST(c.inputFile)
	c.generateSimulationAST(c.outputFile, ast.Labels{"generator": "ast convert", "input_file": c.inputFile})

	return nil
}

func (c *ConvertCommand) loadDslAST(file string) error {
	f, err := os.Open(file)
	if err != nil {
		return fmt.Errorf("Error reading DSL AST file: %v", err)
	}
	defer f.Close()
	c.dslAst, err = ioutil.ReadAll(f)
	return nil
}

func (c *ConvertCommand) generateSimulationAST(file string, labels ast.Labels) error {
	simulation := ast.Simulation{
		Kind: "Simulation",
		Metadata: &ast.ObjectMetadata{
			Labels: &labels,
		},
	}

	arch := gjson.GetBytes(c.dslAst, "object.payload.simulation_arch.value")
	simulation.Spec.Arch = arch.String()

	root := gjson.GetBytes(c.dslAst, "children")

	// Channels
	channelList := buildList(root, "channels", func(value gjson.Result) ast.SimulationChannel {
		simulationChannel := ast.SimulationChannel{
			Name: value.Get("object.payload.channel_name.value").String(),
		}
		// Networks
		networkList := buildList(value, "children.networks", func(value gjson.Result) ast.SimulationNetwork {
			simulationNetwork := ast.SimulationNetwork{
				Name:     value.Get("object.payload.network_name.value").String(),
				MimeType: value.Get("object.payload.mime_type.value").String(),
			}
			return simulationNetwork
		})
		simulationChannel.Networks = &networkList
		return simulationChannel
	})
	simulation.Spec.Channels = channelList

	// Uses
	usesList := buildList(root, "uses", func(value gjson.Result) ast.Uses {
		uses := ast.Uses{
			Name:    value.Get("object.payload.use_item.value").String(),
			Url:     value.Get("object.payload.link.value").String(),
			Version: util.StringPtr(value.Get("object.payload.version.value").String()),
			Path:    util.StringPtr(value.Get("object.payload.path.value").String()),
		}
		return uses
	})
	simulation.Spec.Uses = &usesList

	// Vars
	varsList := buildList(root, "vars", func(value gjson.Result) ast.Var {
		vars := ast.Var{
			Name:  value.Get("object.payload.var_name.value").String(),
			Value: value.Get("object.payload.var_value.value").String(),
		}
		return vars
	})
	simulation.Spec.Vars = &varsList

	// Stacks
	stackList := buildList(root, "stacks", func(value gjson.Result) ast.Stack {
		stack := ast.Stack{
			Name: value.Get("name").String(),
			Arch: func() *string {
				v := value.Get("object.payload.stack_arch.value")
				if v.Exists() == true {
					return util.StringPtr(v.String())
				} else {
					return nil
				}
			}(),
			Stacked: func() *bool {
				v := value.Get("object.payload.stacked.value")
				if v.Exists() == true && v.Bool() == true {
					stacked := v.Bool()
					return &stacked
				} else {
					return nil
				}
			}(),
		}
		// Env
		envList := buildList(value, "env_vars", func(value gjson.Result) ast.Var {
			vars := ast.Var{
				Name:  value.Get("object.payload.env_var_name.value").String(),
				Value: value.Get("object.payload.env_var_value.value").String(),
			}
			return vars
		})
		if len(envList) > 0 {
			stack.Env = &envList
		}
		// Models
		modelList := buildList(value, "children.models", func(value gjson.Result) ast.Model {
			model := ast.Model{
				Name:  value.Get("object.payload.model_name.value").String(),
				Model: value.Get("object.payload.model_repo_name.value").String(),
				Arch: func() *string {
					v := value.Get("object.payload.model_arch.value")
					if v.Exists() == true {
						return util.StringPtr(v.String())
					} else {
						return nil
					}
				}(),
			}
			// Channels
			channelList := buildList(value, "children.channels", func(value gjson.Result) ast.ModelChannel {
				channel := ast.ModelChannel{
					Name:  value.Get("object.payload.channel_name.value").String(),
					Alias: util.StringPtr(value.Get("object.payload.channel_alias.value").String()),
				}
				return channel
			})
			model.Channels = channelList
			// Env
			envList := buildList(value, "children.env_vars", func(value gjson.Result) ast.Var {
				vars := ast.Var{
					Name:  value.Get("object.payload.env_var_name.value").String(),
					Value: value.Get("object.payload.env_var_value.value").String(),
				}
				return vars
			})
			if len(envList) > 0 {
				model.Env = &envList
			}
			// Workflows
			workflowList := buildList(value, "children.workflow", func(value gjson.Result) ast.Workflow {
				workflow := ast.Workflow{
					Name: value.Get("object.payload.workflow_name.value").String(),
				}
				// Vars
				varsList := buildList(value, "children.workflow_vars", func(value gjson.Result) ast.Var {
					vars := ast.Var{
						Name:  value.Get("object.payload.var_name.value").String(),
						Value: value.Get("object.payload.var_value.value").String(),
						// FIXME should uses be decoded here ? and what too ?
					}
					return vars
				})
				workflow.Vars = &varsList
				return workflow
			})
			model.Workflows = &workflowList
			return model
		})
		stack.Models = modelList
		return stack
	})
	simulation.Spec.Stacks = stackList

	if err := util.WriteYaml(&simulation, file, true); err != nil {
		return err
	}
	return nil
}

func buildList[T any](root gjson.Result, match string, gen func(value gjson.Result) T) []T {
	list := []T{}
	matchList := root.Get(match)
	matchList.ForEach(func(key, value gjson.Result) bool {
		item := gen(value)
		list = append(list, item)
		return true
	})
	return list
}

// Copyright 2025 Robert Bosch GmbH
//
// SPDX-License-Identifier: Apache-2.0

package generate

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.boschdevcloud.com/fsil/fsil.go/command/util"
	"github.com/boschglobal/dse.schemas/code/go/dse/ast"
	"github.com/boschglobal/dse.schemas/code/go/dse/kind"
)

func (c *GenerateCommand) GenerateSimulation() error {
	var simulationPath = filepath.Join(c.outputPath, "simulation.yaml")
	os.MkdirAll(filepath.Dir(simulationPath), os.ModePerm)
	os.Remove(simulationPath)
	fmt.Fprintf(flag.CommandLine.Output(), "Writing simulation: %s\n", simulationPath)

	simSpec := c.simulationAst
	//uidList := map[int]interface{}{}
	var currentUid int = 1
	nextUid := func() (uid int) {
		uid = currentUid
		currentUid += 1
		return
	}

	var simbusModel *kind.ModelInstance

	for _, astStack := range simSpec.Stacks {
		stack := kind.Stack{
			Kind: "Stack",
			Metadata: &kind.ObjectMetadata{
				Name: util.StringPtr(astStack.Name),
			},
		}
		configureConnection(&stack)

		// Generate the Models.
		models := []kind.ModelInstance{}
		if simbusModel == nil {
			simbusModel = generateSimbusModel(simSpec)
			models = append(models, *simbusModel)
		}
		for _, astModel := range astStack.Models {
			channels := []kind.Channel{}
			for _, c := range astModel.Channels {
				channels = append(channels, kind.Channel{
					Name:      &c.Name,
					Alias:     &c.Alias,
					Selectors: generateChannelSelectors(astModel, c),
				})
			}

			model := kind.ModelInstance{
				Name: astModel.Name,
				Uid:  nextUid(),
				Model: struct {
					Mcl *struct {
						Models []struct {
							Name string `yaml:"name"`
						} `yaml:"models"`
						Strategy string `yaml:"strategy"`
					} `yaml:"mcl,omitempty"`
					Name string `yaml:"name"`
				}{
					Name: astModel.Model,
				},
				Channels: &channels,
				Runtime:  generateModelRuntime(astModel),
			}
			models = append(models, model)
		}
		stack.Spec.Models = &models

		if err := util.WriteYaml(&stack, simulationPath, true); err != nil {
			return err
		}
	}

	// Write an empty SimBus model.
	simbus := kind.Model{
		Kind: "Model",
		Metadata: &kind.ObjectMetadata{
			Name: util.StringPtr("simbus"),
		},
	}
	if err := util.WriteYaml(&simbus, simulationPath, true); err != nil {
		return err
	}

	return nil
}

func generateChannelSelectors(model ast.Model, channel ast.ModelChannel) *kind.Labels {
	labels := kind.Labels{"model": model.Name}
	if strings.HasSuffix(channel.Alias, "_channel") {
		labels["channel"] = strings.Replace(channel.Alias, "_channel", "_vector", 1)
	}
	return &labels
}

func generateModelRuntime(model ast.Model) *kind.ModelInstanceRuntime {
	env := map[string]string{}
	if model.Env != nil {
		for _, e := range *model.Env {
			env[e.Name] = e.Value
		}
	}
	runtime := kind.ModelInstanceRuntime{
		Env: &env,
		Paths: &[]string{
			fmt.Sprintf("model/%s/data", model.Name),
		},
	}
	return &runtime
}

func configureConnection(stack *kind.Stack) {
	timeout := 60
	redisTransport := kind.StackSpecConnectionTransport0{
		Redis: kind.RedisConnection{
			Timeout: &timeout,
			Uri:     util.StringPtr("redis://localhost:6379"),
		},
	}
	transport := kind.StackSpec_Connection_Transport{}
	transport.FromStackSpecConnectionTransport0(redisTransport)
	connection := struct {
		Timeout   *string                              `yaml:"timeout,omitempty"`
		Transport *kind.StackSpec_Connection_Transport `yaml:"transport,omitempty"`
	}{
		Transport: &transport,
	}
	stack.Spec.Connection = &connection
}

func generateSimbusModel(simSpec ast.SimulationSpec) *kind.ModelInstance {
	channelMap := make(map[string]int)
	for _, channel := range simSpec.Channels {
		channelMap[channel.Name] = 0
	}
	for _, stack := range simSpec.Stacks {
		for _, model := range stack.Models {
			for _, channel := range model.Channels {
				count, ok := channelMap[channel.Name]
				if ok {
					channelMap[channel.Name] = count + 1
				}
			}
		}
	}

	channels := []kind.Channel{}
	for channelName, expectedCount := range channelMap {
		channels = append(channels, kind.Channel{
			Name:               &channelName,
			ExpectedModelCount: &expectedCount,
		})
	}
	model := kind.ModelInstance{
		Name: "simbus",
		Model: struct {
			Mcl *struct {
				Models []struct {
					Name string `yaml:"name"`
				} `yaml:"models"`
				Strategy string `yaml:"strategy"`
			} `yaml:"mcl,omitempty"`
			Name string `yaml:"name"`
		}{
			Name: "simbus",
		},
		Channels: &channels,
	}
	return &model
}

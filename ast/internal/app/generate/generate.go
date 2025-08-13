// Copyright 2025 Robert Bosch GmbH
//
// SPDX-License-Identifier: Apache-2.0

package generate

import (
	"flag"
	"fmt"
	"log/slog"

	"github.com/boschglobal/dse.clib/extra/go/command"
	"github.com/boschglobal/dse.clib/extra/go/command/log"

	"github.com/boschglobal/dse.clib/extra/go/file/handler"
	"github.com/boschglobal/dse.clib/extra/go/file/handler/kind"

	"github.com/boschglobal/dse.schemas/code/go/dse/ast"
)

type GenerateCommand struct {
	command.Command

	inputFile     string
	outputPath    string
	genTaskfile   bool
	genSimulation bool
	logLevel      int

	simulationAst ast.SimulationSpec
}

func NewGenerateCommand(name string) *GenerateCommand {
	c := &GenerateCommand{
		Command: command.Command{
			Name:    name,
			FlagSet: flag.NewFlagSet(name, flag.ExitOnError),
		},
	}
	c.FlagSet().StringVar(&c.inputFile, "input", "", "path to Simulation AST file")
	c.FlagSet().StringVar(&c.outputPath, "output", "", "path to write generated files (Simer layout)")
	c.FlagSet().BoolVar(&c.genTaskfile, "taskfile", false, "Generate a Taskfile (only)")
	c.FlagSet().BoolVar(&c.genSimulation, "simulation", false, "Generate a Simulation (only)")
	c.FlagSet().IntVar(&c.logLevel, "log", 4, "Loglevel")
	return c
}

func (c GenerateCommand) Name() string {
	return c.Command.Name
}

func (c GenerateCommand) FlagSet() *flag.FlagSet {
	return c.Command.FlagSet
}

func (c *GenerateCommand) Parse(args []string) error {
	return c.FlagSet().Parse(args)
}

func (c *GenerateCommand) Run() error {
	var err error
	slog.SetDefault(log.NewLogger(c.logLevel))

	fmt.Fprintf(flag.CommandLine.Output(), "Reading file: %s\n", c.inputFile)
	err = c.loadAst(c.inputFile)
	if err != nil {
		return err
	}

	fmt.Fprintf(flag.CommandLine.Output(), "Writing to folder: %s\n", c.outputPath)

	if c.genTaskfile == false && c.genSimulation == false {
		c.genTaskfile = true
		c.genSimulation = true
	}
	if c.genTaskfile == true {
		err = c.GenerateTaskfile()
		if err != nil {
			return err
		}
	}
	if c.genSimulation == true {
		c.GenerateSimulation()
	}
	return nil
}

func (c *GenerateCommand) loadAst(file string) error {
	_, docs, err := handler.ParseFile(file)
	if err != nil {
		return err
	}
	docList := docs.([]kind.KindDoc)
	for _, doc := range docList {
		slog.Info(fmt.Sprintf("kind: %s; name=%s (%s)", doc.Kind, doc.Metadata.Name, doc.File))
		if doc.Kind == "Simulation" {
			ast := doc.Spec.(*ast.SimulationSpec)
			c.simulationAst = *ast
			return nil
		}
	}
	return fmt.Errorf("Simulation AST not found in file: %s", file)
}

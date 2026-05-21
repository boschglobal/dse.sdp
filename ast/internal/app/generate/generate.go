// Copyright 2025 Robert Bosch GmbH
//
// SPDX-License-Identifier: Apache-2.0

package generate

import (
	"flag"
	"fmt"
	"log/slog"
	"path/filepath"

	"github.com/boschglobal/dse.clib/extra/go/command"
	"github.com/boschglobal/dse.clib/extra/go/command/log"

	"github.com/boschglobal/dse.clib/extra/go/file/handler"
	"github.com/boschglobal/dse.clib/extra/go/file/handler/kind"

	"github.com/boschglobal/dse.schemas/code/go/dse/ast"
)

type GenerateCommand struct {
	command.Command

	inputFile      string
	outputPath     string
	genTaskfile    bool
	genSimulation  bool
	overwriteFiles bool
	dseScriptPath  string
	logLevel       int

	simulationAst ast.SimulationSpec
	simulationDoc *kind.KindDoc
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
	c.FlagSet().BoolVar(&c.overwriteFiles, "overwrite", false, "Overwrite existing embedded files")
	c.FlagSet().StringVar(&c.dseScriptPath, "script", "", "Path to DSE Script file (txtar expansion)")
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

	outDir := "out"
	inputPath := filepath.Join("out", c.inputFile)
	c.inputFile = inputPath
	outputPath := filepath.Join(outDir, c.outputPath)
	c.outputPath = outputPath

	fmt.Fprintf(flag.CommandLine.Output(), "Reading file: %s\n", c.inputFile)
	err = c.loadAst(c.inputFile)
	if err != nil {
		return err
	}

	fmt.Fprintf(flag.CommandLine.Output(), "Writing to folder: %s\n", c.outputPath)

	if err = c.expandDseScriptFiles(); err != nil {
		return err
	}

	if !c.genTaskfile && !c.genSimulation {
		c.genTaskfile = true
		c.genSimulation = true
	}
	if c.genTaskfile {
		if err = c.GenerateTaskfile(); err != nil {
			return err
		}
	}
	if c.genSimulation {
		if err = c.GenerateSimulation(); err != nil {
			return err
		}
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
			c.simulationDoc = &doc
			return nil
		}
	}
	return fmt.Errorf("simulation AST not found in file: %s", file)
}

func (c *GenerateCommand) expandDseScriptFiles() error {
	dseScriptPath := c.dseScriptPath
	if dseScriptPath == "" {
		if c.simulationDoc != nil && c.simulationDoc.Metadata.Labels != nil {
			if p, ok := c.simulationDoc.Metadata.Labels["original_dse_script"]; ok {
				dseScriptPath = p
			}
		}
	}
	if dseScriptPath == "" {
		slog.Info("No DSE script path available, skipping txtar expansion")
		return nil
	}

	outputDir := c.outputPath
	if outputDir == "" {
		outputDir = "."
	}

	slog.Info(fmt.Sprintf("Expanding txtar files from %s into %s", dseScriptPath, outputDir))
	return ExpandTxtar(dseScriptPath, outputDir, c.overwriteFiles)
}

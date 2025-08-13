// Copyright 2025 Robert Bosch GmbH
//
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"flag"
	"log/slog"
	"os"

	"github.com/boschglobal/dse.clib/extra/go/command"
	"github.com/boschglobal/dse.sdp/ast/internal/app/convert"
	"github.com/boschglobal/dse.sdp/ast/internal/app/generate"
	"github.com/boschglobal/dse.sdp/ast/internal/app/resolve"
)

var cmds = []command.CommandRunner{
	command.NewHelpCommand("help"),
	convert.NewConvertCommand("convert"),
	generate.NewGenerateCommand("generate"),
	resolve.NewResolveCommand("resolve"),
}

var usage = `
AST Tools for generating and converting Simulation AST objects/files.

Usage:

    ast <command> [option]

    ast convert -input example/ast.json -output example/ast.yaml
    ast resolve -input example/ast.yaml
    ast generate -input example/ast.yaml -output example/sim

`

func printUsage() {
	command.PrintUsage(usage[1:], cmds)
}

func main() {
	os.Exit(main_())
}

func main_() int {
	flag.Usage = printUsage
	if len(os.Args) == 1 {
		printUsage()
		return 1
	}
	if err := command.DispatchCommand(os.Args[1], cmds); err != nil {
		slog.Error(err.Error())
		return 2
	}

	return 0
}

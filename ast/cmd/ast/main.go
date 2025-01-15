// Copyright 2025 Robert Bosch GmbH
//
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"flag"
	"log/slog"
	"os"

	"github.boschdevcloud.com/fsil/fsil.go/command"
	"github.com/boschglobal/dse.sdp/ast/internal/app/convert"
)

var cmds = []command.CommandRunner{
	command.NewHelpCommand("help"),
	convert.NewConvertCommand("convert"),
}

var usage = `
AST Tools

	ast <command> [command options,]
`

func printUsage() {
	command.PrintUsage(usage, cmds)
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

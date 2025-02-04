// Copyright 2024 Robert Bosch GmbH
//
// SPDX-License-Identifier: Apache-2.0
// SPDX-License-Identifier: BSAL-1.0

package main

import (
	"flag"
	"log/slog"
	"os"

	"github.boschdevcloud.com/fsil/fsil.go/command"
	"github.com/boschglobal/dse.sdp/graph/internal/app/graph"
)

var cmds = []command.CommandRunner{
	command.NewHelpCommand("help"),
	graph.NewGraphImportCommand("import"),
	graph.NewGraphExportCommand("export"),
	graph.NewGraphDropCommand("drop"),
}

var usage = `
Graph Tools for importing files to database and reporting.

Usage:

  graph <command> [command options,]

`

func main() {
	os.Exit(main_())
}

func main_() int {
	flag.Usage = PrintUsage
	if len(os.Args) == 1 {
		PrintUsage()
		return 1
	}
	// Dispatch the command.
	if err := command.DispatchCommand(os.Args[1], cmds); err != nil {
		slog.Error(err.Error())
		return 2
	}

	return 0
}

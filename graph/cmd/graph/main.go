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
	graph.NewGraphReportCommand("report"),
}

var usage = `
Graph Tools for importing files to database and reporting.

Usage:

  graph <command> [command options,]
  graph report [--tag=name --db=db_uri] <report file>

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
	// Dispatch the command.
	if err := command.DispatchCommand(os.Args[1], cmds); err != nil {
		slog.Error(err.Error())
		return 2
	}

	return 0
}

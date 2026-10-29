// Copyright 2026 Robert Bosch GmbH
//
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"fmt"
	"maps"
	"os"
	"os/exec"
	"slices"
	"strings"
)

type command struct {
	name    string
	handler func([]string) error
	usage   string
	help    string
}

var commands = map[string]command{
	"validate": {name: "validate", handler: runValidate,
		usage: "[--format text|json] <script.dse>|-",
		help:  "Check syntax and emit text or JSON diagnostics; text is the default."},
	"ebnf": {name: "ebnf", handler: runEbnf,
		help: "Print the DSE Script grammar in EBNF."},
	"compile": {name: "compile", handler: runCompile,
		usage: "[--out output.json|-] <script.dse>|-",
		help:  "Compile a DSE script and write its AST to stdout by default."},
}

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	selected, ok := commands[os.Args[1]]
	if !ok {
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n", os.Args[1])
		printUsage()
		os.Exit(1)
	}

	if err := selected.handler(os.Args[2:]); err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			os.Exit(exitError.ExitCode())
		}
		fmt.Fprintf(os.Stderr, "Launcher failed to execute command %q: %v\n", selected.name, err)
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Fprint(os.Stderr, formatUsage())
}

func formatUsage() string {
	commandNames := slices.Sorted(maps.Keys(commands))
	var output strings.Builder
	fmt.Fprintln(&output, "DSE Lang")
	fmt.Fprintln(&output, "  Containerized DSE Script compiler and validator.")
	fmt.Fprintln(&output)
	fmt.Fprintln(&output, "Usage:")
	fmt.Fprintln(&output, "  docker run [docker options] dse-lang:test <command> [arguments]")
	fmt.Fprintln(&output)
	fmt.Fprintln(&output, "Examples:")
	fmt.Fprintln(&output, "  cat script.dse | docker run --rm -i dse-lang:test validate -")
	fmt.Fprintln(&output, "  docker run --rm -i --volume \"$PWD:/work\" --workdir /work dse-lang:test compile --out ast.json - < script.dse")
	fmt.Fprintln(&output, "  docker run --rm dse-lang:test ebnf")
	fmt.Fprintln(&output)
	fmt.Fprintln(&output, "Commands:")
	for index, name := range commandNames {
		if index > 0 {
			fmt.Fprintln(&output)
		}
		command := commands[name]
		fmt.Fprintf(&output, "  %s", name)
		if command.usage != "" {
			fmt.Fprintf(&output, " %s", command.usage)
		}
		fmt.Fprintln(&output)
		fmt.Fprintf(&output, "    %s\n", command.help)
	}
	return output.String()
}

func runValidate(args []string) error {
	return runCompiler(normalizeValidateArgs(args))
}

func runCompile(args []string) error {
	return runCompiler(normalizeCompileArgs(args))
}

func runCompiler(args []string) error {
	nodeArgs := append([]string{"/app/dse-lang-compile.cjs"}, args...)
	cmd := exec.Command("/nodejs/bin/node", nodeArgs...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func normalizeValidateArgs(args []string) []string {
	normalized := make([]string, 0, len(args)+2)
	hasFormat := false

	for i := 0; i < len(args); i++ {
		arg := args[i]
		if arg == "--ast" {
			if i+1 < len(args) && !strings.HasPrefix(args[i+1], "--") {
				i++
			}
			continue
		}
		if strings.HasPrefix(arg, "--ast=") {
			continue
		}
		if arg == "--format" || strings.HasPrefix(arg, "--format=") {
			hasFormat = true
		}
		normalized = append(normalized, arg)
	}

	if !hasFormat {
		normalized = append([]string{"--format", "text"}, normalized...)
	}
	return normalized
}

func normalizeCompileArgs(args []string) []string {
	normalized := make([]string, 0, len(args)+2)
	hasAst := false
	for i := 0; i < len(args); i++ {
		arg := args[i]
		if arg == "--" {
			if !hasAst {
				normalized = append(normalized, "--ast")
			}
			return append(normalized, args[i:]...)
		}
		if arg == "--format" {
			if i+1 < len(args) && !strings.HasPrefix(args[i+1], "--") {
				i++
			}
			continue
		}
		if strings.HasPrefix(arg, "--format=") {
			continue
		}
		if arg == "--out" {
			hasAst = true
			normalized = append(normalized, "--ast")
			if i+1 < len(args) && !strings.HasPrefix(args[i+1], "--") {
				i++
				normalized = append(normalized, args[i])
			}
			continue
		}
		if strings.HasPrefix(arg, "--out=") {
			hasAst = true
			if arg == "--out=" {
				normalized = append(normalized, "--ast")
			} else {
				normalized = append(normalized, "--ast="+strings.TrimPrefix(arg, "--out="))
			}
			continue
		}
		if arg == "--ast" {
			hasAst = true
			normalized = append(normalized, arg)
			if i+1 < len(args) && !strings.HasPrefix(args[i+1], "--") {
				i++
				normalized = append(normalized, args[i])
			}
			continue
		}
		if strings.HasPrefix(arg, "--ast=") {
			hasAst = true
			if arg == "--ast=" {
				normalized = append(normalized, "--ast", "-")
			} else {
				normalized = append(normalized, arg)
			}
			continue
		}
		normalized = append(normalized, arg)
	}
	if !hasAst {
		normalized = append(normalized, "--ast")
	}
	return normalized
}

func runEbnf(_ []string) error {
	ebnfPath := "/usr/share/dse-lang/dse-script.ebnf"
	content, err := os.ReadFile(ebnfPath)
	if err != nil {
		fmt.Fprintf(os.Stdout, "Unable to read EBNF file %q: %v\n", ebnfPath, err)
		return err
	}
	if _, err := os.Stdout.Write(content); err != nil {
		return err
	}
	return nil
}

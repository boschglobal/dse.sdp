// Copyright 2026 Robert Bosch GmbH
//
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"reflect"
	"testing"
)

func TestFormatUsage(t *testing.T) {
	want := "DSE Lang\n" +
		"  Containerized DSE Script compiler and validator.\n\n" +
		"Usage:\n" +
		"  docker run [docker options] dse-lang:test <command> [arguments]\n\n" +
		"Examples:\n" +
		"  cat script.dse | docker run --rm -i dse-lang:test validate -\n" +
		"  docker run --rm -i --volume \"$PWD:/work\" --workdir /work dse-lang:test compile --out ast.json - < script.dse\n" +
		"  docker run --rm dse-lang:test ebnf\n\n" +
		"Commands:\n" +
		"  compile [--out output.json|-] <script.dse>|-\n" +
		"    Compile a DSE script and write its AST to stdout by default.\n\n" +
		"  ebnf\n" +
		"    Print the DSE Script grammar in EBNF.\n\n" +
		"  validate [--format text|json] <script.dse>|-\n" +
		"    Check syntax and emit text or JSON diagnostics; text is the default.\n"
	if got := formatUsage(); got != want {
		t.Errorf("formatUsage() = %q, want %q", got, want)
	}
}

func TestNormalizeValidateArgs(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want []string
	}{
		{
			name: "default format and stdin",
			args: []string{"-"},
			want: []string{"--format", "text", "-"},
		},
		{
			name: "preserve format and remove ast output",
			args: []string{"--format", "json", "--ast", "ast.json", "input.dse"},
			want: []string{"--format", "json", "input.dse"},
		},
		{
			name: "equals forms",
			args: []string{"--format=json", "--ast=ast.json", "-"},
			want: []string{"--format=json", "-"},
		},
		{
			name: "ast without value does not remove next option",
			args: []string{"--ast", "--format", "json", "-"},
			want: []string{"--format", "json", "-"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := normalizeValidateArgs(test.args); !reflect.DeepEqual(got, test.want) {
				t.Errorf("normalizeValidateArgs(%q) = %q, want %q", test.args, got, test.want)
			}
		})
	}
}

func TestNormalizeCompileArgs(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want []string
	}{
		{
			name: "default ast output to stdout and remove format",
			args: []string{"--format", "json", "-"},
			want: []string{"-", "--ast"},
		},
		{
			name: "preserve explicit ast output and remove format",
			args: []string{"--format", "json", "--ast", "ast.json", "-"},
			want: []string{"--ast", "ast.json", "-"},
		},
		{
			name: "preserve bare ast option",
			args: []string{"--ast"},
			want: []string{"--ast"},
		},
		{
			name: "remove equals format",
			args: []string{"--format=text", "input.dse"},
			want: []string{"input.dse", "--ast"},
		},
		{
			name: "map out path to ast path",
			args: []string{"--out", "ast.json", "input.dse"},
			want: []string{"--ast", "ast.json", "input.dse"},
		},
		{
			name: "map equals out path to ast path",
			args: []string{"--out=ast.json", "input.dse"},
			want: []string{"--ast=ast.json", "input.dse"},
		},
		{
			name: "preserve explicit equals ast output",
			args: []string{"--ast=ast.json", "input.dse"},
			want: []string{"--ast=ast.json", "input.dse"},
		},
		{
			name: "default ast before option terminator",
			args: []string{"--", "--format", "input.dse"},
			want: []string{"--ast", "--", "--format", "input.dse"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := normalizeCompileArgs(test.args); !reflect.DeepEqual(got, test.want) {
				t.Errorf("normalizeCompileArgs(%q) = %q, want %q", test.args, got, test.want)
			}
		})
	}
}

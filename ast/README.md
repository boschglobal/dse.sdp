<!--
Copyright 2025 Robert Bosch GmbH

SPDX-License-Identifier: Apache-2.0
-->

# DSE SDP AST Tools

AST Tools and Library.


## Usage

```bash
# Build.
$ make

# Build the DSL parser (optional).
$ cd ../dsl
$ make
$ cd -

# AST Tools (including parsing the DSL).
$ parse2ast simulation.dse simulation.ast.json
$ bin/ast convert -input simulation.ast.json -output ast.yaml
```


## Development

### Go Module Update (schema updates)

```bash
$ export GOPRIVATE=github.com/boschglobal,github.boschdevcloud.com
$ go clean -modcache
$ go mod tidy

# Go get (adjust version as required).
$ go get -x github.com/boschglobal/dse.schemas/code/go/dse@v1.2.24
$ go get github.com/boschglobal/dse.clib/extra/go/command
```

> Note: Release Tags for modules in DSE Schemas are according to the schema `code/go/dse/v1.2.24`.

> Note: Release Tags for modules in FSIL Go are according to the schema `command/v1.0.5`.


### Generating Test Data

```bash
$ cd dsl
$ make
$ parse2ast ../ast/cmd/ast/testdata/dsl/single_model.dse ../ast/cmd/ast/testdata/dsl/single_model.ast.json
```

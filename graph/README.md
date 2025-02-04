<!--
Copyright 2025 Robert Bosch GmbH

SPDX-License-Identifier: Apache-2.0
-->

# DSE SDP Graph Tools

Graph Tools and Library.


## Usage

```bash
# Build.
$ make

# Start a Graph (if not running).
$ make graph

# Run tests.
$ make test
$ make test_e2e

# Graph Tools
# Drop all nodes from graph.
$ bin/graph drop --all
Running command graph ...
2025/01/27 15:01:17 INFO Connect to graph db=bolt://localhost:7687
Graph query: MATCH DETACH DELETE all nodes

# Import files to the database.
$ bin/graph import internal/pkg/file/kind/testdata/brake-by-wire/simulation.yaml
Running command graph ...
2025/01/28 09:21:50 INFO Connect to graph db=bolt://localhost:7687
  Handler:  yaml/kind=Stack
  ...

# Export the database to a file.
$ bin/graph export export.cyp
Running command graph ...
2025/01/28 09:25:45 INFO Connect to graph db=bolt://localhost:7687
2025/01/28 09:25:45 INFO Graph Export file=export.cyp
2025/01/28 09:25:45 INFO Graph QueryRecord query="CALL export_util.cypher_all(\"\", {stream: true}) YIELD data RETURN data" params=map[]
2025/01/28 09:25:45 INFO Graph Export: write file=export.cyp
...

# Graph is available at:
http://localhost:3000/lab/dashboard?component=query
#  Query: MATCH (node1)-[r*]->(node2) RETURN node1, r, node2;

```


## Development

### Go Module Update (schema updates)

```bash
$ export GOPRIVATE=github.com/boschglobal,github.boschdevcloud.com
$ go clean -modcache
$ go mod tidy

# Go get (adjust version as required).
$ go get -x github.com/boschglobal/dse.schemas/code/go/dse@v1.2.13
$ go get go get github.com/stretchr/testify
$ go get github.com/stretchr/testify/assert@v1.10.0
$ go get github.com/gabriel-vasile/mimetype
$ go get github.com/neo4j/neo4j-go-driver/v5@v5.22.0
$ go get gopkg.in/yaml.v3@v3.0.1
$ go get github.com/oapi-codegen/runtime@v1.1.1
$ go get github.com/rogpeppe/go-internal/testscript
```

> Note: Release Tags for modules in DSE Schemas are according to the schema `code/go/dse/v1.2.11`.



### Go Module Vendor

```bash
# Vendor the project.
$ go mod tidy
$ go mod vendor
```

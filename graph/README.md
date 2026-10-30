<!--
Copyright 2025 Robert Bosch GmbH

SPDX-License-Identifier: Apache-2.0
-->

# DSE SDP Graph Tools


## Introduction

The Graph Tools import simulation configuration files (YAML) into a graph
database ([Memgraph](https://memgraph.com/)), and run reports (Cypher queries)
against the resulting graph. The code is also a library for working with the
graph from other tools.


### Contents

- [Introduction](#introduction)
  - [Project Structure](#project-structure)
- [Usage](#usage)
  - [Logging](#logging)
- [Development](#development)
  - [Build](#build)
  - [Test](#test)
  - [Dependencies](#dependencies)


### Project Structure

```text
graph/
├── build/package/report/       # Report container image
├── cmd/graph/                  # CLI entry point
│   ├── reports/                # Report definitions (YAML)
│   └── testdata/               # End-to-end test scripts
└── internal/
    ├── app/graph/              # CLI commands (import, export, drop, report, ping)
    └── pkg/
        ├── file/               # File type detection and import handlers
        │   └── kind/           # Graph mapping for each YAML kind
        └── graph/              # Graph database access
            └── graphtest/      # Memgraph fixture for tests
```


## Usage

Start Memgraph, then use the CLI:

```bash
# Start Memgraph (if not running), and build the CLI.
make graph
make build

# Remove nodes: all, or only 'ast' or 'sim' nodes.
bin/graph drop --all

# Import a file, or a folder of files.
bin/graph import ../examples/graph/static_validation/sim_good

# Run the reports (filtered by tag) against the imported simulation.
bin/graph report --reports cmd/graph/reports --tag stack ../examples/graph/static_validation/sim_good

# List the available reports.
bin/graph report --reports cmd/graph/reports --list

# Export the graph as Cypher.
bin/graph export export.cyp
```

Add `--db` to any command to use a database other than `bolt://localhost:7687`.
Credentials are read from the `GRAPH_USER` and `GRAPH_PASSWORD` environment
variables (none by default).

Explore the graph at <http://localhost:3000> (Memgraph Lab), for example with
`MATCH (a)-[r*]->(b) RETURN a, r, b;`.


### Logging

All commands accept `--log <level>` and write log output to stderr. Use it to
trace the Cypher queries sent to the graph and the responses received.

| Level | Name | Output |
| :--- | :--- | :--- |
| `4` | error | Errors only (default) |
| `3` | warn | Warnings |
| `2` | info | Progress, such as the files imported |
| `1` | debug | Each Cypher query, and the summary of its response |
| `0` | trace | All Bolt messages: query parameters, every record returned, and driver internals |

Debug shows what was sent, and whether it succeeded:

```bash
bin/graph drop --all --log 1
...
DEBUG: Graph client conn=bolt-1@localhost:7687 msg=BEGIN {}
DEBUG: Graph client conn=bolt-1@localhost:7687 msg=RUN "MATCH (n) DETACH DELETE n"
DEBUG: Graph server conn=bolt-1@localhost:7687 msg=SUCCESS {"db":"memgraph"}
DEBUG: Graph client conn=bolt-1@localhost:7687 msg=COMMIT
```

Trace adds the query parameters and every record. It prints a lot, so save it
to a file and search it:

```bash
bin/graph report --log 0 --reports cmd/graph/reports ../examples/graph/static_validation/sim_good 2> trace.log
grep 'msg=RUN' trace.log     # Queries with parameters.
grep 'msg=RECORD' trace.log  # Records returned.
```

Query parameters and records are logged at trace level only, as they can hold
sensitive values (for example repository tokens in a Manifest). The driver
redacts credentials from its own messages. Trace is shown as `DEBUG-4` in the
output.

In tests, set a logger on `slog` with `graph.LevelTrace` to capture the same
output.


## Development

### Build

```bash
git clone https://github.com/boschglobal/dse.sdp.git
cd dse.sdp/graph

# Build
make build
make docker

# Install the CLI and reports to ~/.local
make install

# Additional Makefile targets
make help
```


### Test

Tests use the Memgraph fixture in `internal/pkg/graph/graphtest`. The fixture
selects the database as follows:

- `GRAPH_DB=bolt://host:7687` uses an existing instance (for example a CI service container).
- Otherwise a temporary container is started with docker (`GRAPH_IMAGE` overrides the image).
- Without either, tests skip locally, and fail when `CI` is set.

```bash
# Unit tests (with the race detector), a temporary Memgraph is started automatically.
make test

# End-to-end tests, these use the Memgraph started by `make graph`.
make test_e2e
make graph-stop

# Format and lint (golangci-lint in docker, see .golangci.yml).
make lint
```


### Dependencies

```bash
export GOPRIVATE=github.com/boschglobal
go get -u ./...
go mod tidy
```

> Note: Release tags for modules in DSE Schemas follow the scheme `code/go/dse/v1.2.11`.

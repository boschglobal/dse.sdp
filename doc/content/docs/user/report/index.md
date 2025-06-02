---
title: "Graph - Report Appliance"
linkTitle: "Report"
weight: 20
tags:
- Graph
- Report
- CLI
github_repo: "https://github.boschdevcloud.com/fsil/dse.sdp"
github_subdir: "doc"
---


## Synopsis

Containerized report appliance that runs the installed reports on specified simulation package.

```bash
# Run the reports.
$ dse-report examples/graph/<sim_name>/<simulation_status>
```

## Simulation Setup

### Structure

The structure of a simulation follows the format examples/graph/<sim_name>/<simulation_status> # where simulation_status is sim_good or sim_with_error.

The graph-report command will run all installed reports on the specified simulation package.

### Simulation Layout

```text
examples/
└── graph/
    └── duplicate_writes/
        ├── sim_good/
        │   ├── data/
        │   │   └── simulation.yaml
        │   └── model/
        │       ├── input/
        │       │   └── data/
        │       │       ├── model.yaml
        │       │       └── signalgroup.yaml
        │       └── linear/
        │           └── data/
        │               ├── model.yaml
        │               └── signalgroup.yaml
        └── sim_with_error/
            ├── data/
            │   └── simulation.yaml
            └── model/
                ├── input/
                │   └── data/
                │       ├── model.yaml
                │       └── signalgroup.yaml
                └── linear/
                    └── data/
                        ├── model.yaml
                        └── signalgroup.yaml
          stack/
          ├── sim_good/
          │   └── simulation.yaml
          └── sim_with_error/
              └── simulation.yaml

          static_validation/
          └── sim_with_error/
              └── data/
                  └── simulation.yaml
```

## Report Appliance

### Setup

Report appliance is a containerized tool that runs cypher queries defined in YAML reports, evaluates their outcomes, and logs pass/fail results for simulation model validations.

```bash
# Start the memgraph container (using a make target).
$ make graph

# Build containerized graph tool.
$ make docker
```

#### Shell Function

```bash
# Define a shell function for the report command.
$ dse-report() {
  docker run -t --rm \
    -v memgraph:/var/lib/memgraph \
    -v "$(pwd)":/sim \
    dse-graph:test "$@"
}

# Run the reports.
$ dse-report examples/graph/static_validation/sim_good
```


### Options

```bash
$ dse-report -h
Running command: report
Usage of report:
  -db string
        database connection string (default "bolt://localhost:7687")
  -list
        list all available reports and their tags
  -list-all
        list all available report details in tabular format
  -list-tags
        list all available tags from reports
  -name value
        run report with specified report name(s)
  -reports string
        run all reports form the specified reports folder
  -tag value
        run all reports with specified tag
```

## Definitions


Cypher
: Declarative query language used to interact with graph databases.

Report
: A YAML file that defines database queries and expected results, to check correctness of imported simulation data.

Simulation
: The structural arrangement of several Models, which, when connected via a Simulation Bus, form a simulation. Typically defined in a file `simulation.yaml`.

Simulation Package (simulation path/folder)
: This folder contains all configuration, models and artifacts necessary to run a simulation.

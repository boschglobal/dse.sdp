---
title: "Report - Simulation Validation"
linkTitle: "Report"
weight: 40
tags:
- Graph
- Report
- CLI
github_repo: "https://github.com/boschglobal/dse.sdp"
github_subdir: "doc"
---


## Synopsis

Containerized simulation validation tool for Simer based simulations.

```bash
# Run the reports.
$ dse-report path/to/simulation
```


## Report Tool

### Codespace

The Codespace (aka Devcontainer) of the DSE Simulation Development Platform is pre-configured with a command for running the Report tool.

```bash
# Run the reports.
$ dse-report path/to/simulation
```


### Setup

Report is a containerized tool which validates a simulation using a collection of report templates included in the Report container.

```bash
# Latest Report Container:
$ docker pull ghcr.io/boschglobal/dse-report:latest

# Specific versions of the Report Container
$ docker pull ghcr.io/boschglobal/dse-report:0.1.1
$ docker pull ghcr.io/boschglobal/dse-simer:0.1
```

#### Shell Function

```bash
# Define a shell function for the report command.
$ export DSE_REPORT_IMAGE=ghcr.io/boschglobal/dse-report:latest
$ dse-report() { ( \
      if test -d "$1"; then cd "$1" && shift; fi && \
      docker run -t --rm \
      -v $(pwd):/sim \
      $DSE_REPORT_IMAGE /sim "$@";
) }

# Run the reports.
$ dse-report path/to/simulation
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


## Examples

### Included Reports

Run the Report Tool with all included reports.

```bash
$ dse-report examples/graph/stack/sim_good
...
=== Summary ===================================================================
[PASS] Duplicate Writes Check
[PASS] ModelInstance Name Check
[PASS] Model UID Check
[PASS] Channel 'expectedModelCount'
[PASS] Count 'ModelInst' in AST and SIM
Ran 5 Reports | Passed: 5 | Failed: 0
```

<details>
<summary>Report Output</summary>

```bash
$ dse-report examples/graph/stack/sim_good
Running command: report
Options:
  db             : bolt://localhost:7687
  list           : false
  list-all       : false
  list-tags      : false
  log            : 4
  name           : 
  reports        : 
  tag            : 

=== Files ===================================================================
sim/simulation.yaml

=== Report ===================================================================
Name: Duplicate Writes Check
Path: /home/memgraph/.local/share/dse-graph/reports/duplicate_writes.yaml
Version: 0.0.0
Date: 2025-06-02 12:55:03
Query: Duplicate Write Signals
Cypher:
    // Get output signals for each SimbusChannel.
    MATCH (sc:SimbusChannel)<-[:Belongs]-(ch1:Channel)
    -[:Represents]->(:SignalGroup)-[:Contains]->(s1:Signal)
    WHERE s1.annotations.fmi_variable_causality = "output"
    WITH sc AS simbus_channel, collect(DISTINCT s1.name) AS output_signals
    
    // Get input signals from input model instance with matching selector.channel.
    MATCH (sc:SimbusChannel)<-[:Belongs]-(ch2:Channel)<-[:Alias]-
    (mi:ModelInst {name: "input"})-[:InstanceOf]->(:Model)
    MATCH (mi)-[:Has]->(sel:Selector)-[:Selects]->(:Label)
    <-[:Has]-(:SignalGroup)-[:Contains]->(s2:Signal)
    WITH simbus_channel, output_signals, collect(DISTINCT s2.name) AS input_signal_list
    
    // Find intersection using UNWIND and WHERE.
    UNWIND input_signal_list AS individual_input_signal
    WITH simbus_channel, output_signals, input_signal_list, individual_input_signal
    WHERE individual_input_signal IN output_signals
    WITH simbus_channel, output_signals, input_signal_list, 
    collect(individual_input_signal) AS common_signals
    
    RETURN simbus_channel, common_signals
Results:
Evaluation: Report Passed

=== Report ===================================================================
Name: ModelInstance Name Check
Path: /home/memgraph/.local/share/dse-graph/reports/stack.yaml
Version: 0.0.0
Date: 2025-06-02 12:55:03
Query: Unique ModelInstance Name
Cypher:
    MATCH (:Stack)-[:Has]->(mi:ModelInst)
    WHERE mi.name IS NOT NULL
    WITH mi.name AS name, count(*) AS count
    WHERE count > 1
    RETURN name, count
Results:
Evaluation: Report Passed

=== Report ===================================================================
Name: Model UID Check
Path: /home/memgraph/.local/share/dse-graph/reports/stack.yaml
Version: 0.0.0
Date: 2025-06-02 12:55:03
Query: Unique Non-zero Model UID
Cypher:
    MATCH (:Stack)-[:Has]->(mi:ModelInst)
    WHERE mi.uid IS NOT NULL
    WITH mi.uid AS uid, collect(mi.name) AS names, count(*) AS count
    WHERE uid = "0" OR count > 1
    RETURN uid, count, names
Results:
Evaluation: Report Passed

=== Report ===================================================================
Name: Channel 'expectedModelCount'
Path: /home/memgraph/.local/share/dse-graph/reports/static_validation.yaml
Version: 0.0.0
Date: 2025-06-02 12:55:03
Query: Expected Count
Cypher:
    MATCH (mi:ModelInst)-[:Alias]->(ch:Channel)
    WITH ch.name AS channelName, COUNT(DISTINCT mi) AS actualCount
    MATCH (:Stack)-[:Has]->(:Simbus)-[:Has]->(sc:SimbusChannel)
    WHERE sc.name = channelName
    RETURN channelName,
          sc.expectedModelCount AS expectedCount,
          actualCount,
          CASE WHEN sc.expectedModelCount = actualCount THEN "PASS" ELSE "FAIL" END AS result
Results:
+--------------+---------------+-------------+--------+
| CHANNELNAME  | EXPECTEDCOUNT | ACTUALCOUNT | RESULT |
+--------------+---------------+-------------+--------+
| data_channel |             4 |           4 | PASS   |
+--------------+---------------+-------------+--------+
Evaluation: Report Passed
Query: Model to Channel Mapping
Cypher:
    MATCH (st:Stack)-[:Has]->(mi:ModelInst)-[a:Alias]->(ch:Channel)
    WITH mi, a, ch
    RETURN mi.name AS modelInstName, a.name as alias, ch.name AS channelName
Results:
+---------------+-------+--------------+
| MODELINSTNAME | ALIAS | CHANNELNAME  |
+---------------+-------+--------------+
| counter_A     | data  | data_channel |
| counter_B     | data  | data_channel |
| counter_C     | data  | data_channel |
| counter_D     | data  | data_channel |
+---------------+-------+--------------+

=== Report ===================================================================
Name: Count 'ModelInst' in AST and SIM
Path: /home/memgraph/.local/share/dse-graph/reports/static_validation.yaml
Version: 0.0.0
Date: 2025-06-02 12:55:03
Query: Expected Count
Cypher:
    MATCH (fl:File)-[:Contains]->(st:Stack)-[:Has]->(mi:ModelInst)
    WITH fl, COUNT(DISTINCT mi) AS countSim
    MATCH (fl)-[:Contains]->(sim:Simulation)-[:Has]->(st2:Stack)-[:Has]->(mi2:ModelInst)
    WITH countSim, COUNT(DISTINCT mi2) AS countAst
    RETURN
        countAst AS astModelInstCount,
        countSim AS simModelInstCount,
        CASE WHEN countAst = countSim THEN "PASS" ELSE "FAIL" END AS result
Results:
No records found
Evaluation: Report Passed

=== Summary ===================================================================
[PASS] Duplicate Writes Check
[PASS] ModelInstance Name Check
[PASS] Model UID Check
[PASS] Channel 'expectedModelCount'
[PASS] Count 'ModelInst' in AST and SIM
Ran 5 Reports | Passed: 5 | Failed: 0
```
</details>


## Definitions

Cypher
: Declarative query language used to interact with graph databases.

Report
: A YAML file that defines database queries and expected results, to check correctness of imported simulation data.

Simulation
: The structural arrangement of several Models, which, when connected via a Simulation Bus, form a simulation. Typically defined in a file `simulation.yaml`.

Simulation Package (simulation path/folder)
: This folder contains all configuration, models and artifacts necessary to run a simulation.

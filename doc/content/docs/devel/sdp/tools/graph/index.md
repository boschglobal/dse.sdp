---
title: "Graph - Graph Tools"
linkTitle: "Graph"
weight: 100
tags:
- SDP
- CLI
github_repo: "https://github.com/boschglobal/dse.sdp"
github_subdir: "doc"
---


## Synopsis
Graph Tools for static and dynamic analysis of Simulations.

```bash
$ dse-graph report examples/graph/<sim-name>/<sim-status>
```


## Commands
The Graph tool includes the following commands and options:


### Drop

```bash
$ dse-graph drop <sim|ast|-all>
```


### Import

```bash
dse-graph import examples/graph/<sim-name>/<sim-status>
```


### Export

```bash
$ dse-graph export export.cyp
```


### Report

```bash
dse-graph report examples/graph/<sim-name>/<sim-status>
```
#### Option Tag (-tag)

```bash
dse-graph report -tag=tag_name examples/graph/<sim-name>/<sim-status>
```
#### Option List (-list)

```bash
dse-graph report <-list|-list-all|-list-tags>
```
#### Option Name (-name)

```bash
dse-graph report -name="Report1;Report2" examples/graph/<sim-name>/<sim-status>
```
#### Option Reports (-reports)

```bash
dse-graph report -reports=path/to/reports examples/graph/<sim-name>/<sim-status>
```


## Report Appliance

### Usage

```bash
# Start the memgraph container (using a make target).
$ make graph

# Build containerized graph tool.
$ make docker

# Run tests.
$ make test

# Define a shell function for the report command.
$ dse-report() {
  docker run -t --rm \
    -v "$(pwd)":/sim \
    dse-report:test "$@"
}

# Run the reports on a simulation.
$ dse-report examples/graph/static_validation/sim_good
...
=== Summary ===================================================================
[PASS] Duplicate Writes Check
[PASS] ModelInstance Name Check
[PASS] Model UID Check
[PASS] Channel 'expectedModelCount'
[PASS] Count 'ModelInst' in AST and SIM
Ran 5 Reports | Passed: 5 | Failed: 0

# Run specific reports by name.
$ dse-report --name="Model UID Check;ModelInstance Name Check" examples/graph/stack/sim_good
...
=== Summary ===================================================================
[PASS] ModelInstance Name Check
[PASS] Model UID Check
Ran 2 Reports | Passed: 2 | Failed: 0
```

## Reports

### Static Validation

Performs a collection of static validation checks on simulation configuration files.

<details>
<summary>Static Validation - Simulation with no errors</summary>

```bash
$ dse-report --name="Channel 'expectedModelCount';Count 'ModelInst' in AST and SIM" examples/graph/static_validation/sim_good
...
Pinging Memgraph...
...
Running command: drop
...
Running command: report
Options:
  db             : bolt://localhost:7687
  list           : false
  list-all       : false
  list-tags      : false
  name           :
  reports        :
  tag            :
2025/05/22 20:05:59 INFO Connect to graph db=bolt://localhost:7687
  Handler:  yaml/kind=Stack
  Handler:  yaml/kind=Model
  Handler:  yaml/kind=Simulation
simulation.yaml
Stack
simulation.yaml
Model
simulation.yaml
Simulation

=== Report ===================================================================
Name: Channel 'expectedModelCount'
Path: /home/memgraph/.local/share/dse-graph/reports/static_validation.yaml
Version: 0.0.0
Date: 2025-05-27 10:14:51
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
+-------------+---------------+-------------+--------+
| CHANNELNAME | EXPECTEDCOUNT | ACTUALCOUNT | RESULT |
+-------------+---------------+-------------+--------+
| Network     |             1 |           1 | PASS   |
| physical    |             2 |           2 | PASS   |
+-------------+---------------+-------------+--------+
Evaluation: Report Passed
Query: Model to Channel Mapping
Cypher:
    MATCH (st:Stack)-[:Has]->(mi:ModelInst)-[a:Alias]->(ch:Channel)
    WITH mi, a, ch
    RETURN mi.name AS modelInstName, a.name as alias, ch.name AS channelName

Results:
+---------------+-----------------+-------------+
| MODELINSTNAME | ALIAS           | CHANNELNAME |
+---------------+-----------------+-------------+
| input         | scalar          | physical    |
| linear        | signal_channel  | physical    |
| linear        | network_channel | Network     |
+---------------+-----------------+-------------+


=== Report ===================================================================
Name: Count 'ModelInst' in AST and SIM
Path: /home/memgraph/.local/share/dse-graph/reports/static_validation.yaml
Version: 0.0.0
Date: 2025-05-27 10:14:51
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
+-------------------+-------------------+--------+
| ASTMODELINSTCOUNT | SIMMODELINSTCOUNT | RESULT |
+-------------------+-------------------+--------+
|                 2 |                 2 | PASS   |
+-------------------+-------------------+--------+
Evaluation: Report Passed


=== Summary ===================================================================
[PASS] Channel 'expectedModelCount'
[PASS] Count 'ModelInst' in AST and SIM
Ran 2 Reports | Passed: 2 | Failed: 0
```
</details>


<details>
<summary>Static Validation - Simulation <b>with</b> errors</summary>

```bash
$ dse-report --name="Channel 'expectedModelCount';Count 'ModelInst' in AST and SIM" examples/graph/static_validation/sim_with_error
...
Pinging Memgraph...
...
Running command: drop
...
Running command: report
Options:
  db             : bolt://localhost:7687
  list           : false
  list-all       : false
  list-tags      : false
  name           :
  reports        :
  tag            :
2025/05/22 20:10:03 INFO Connect to graph db=bolt://localhost:7687
  Handler:  yaml/kind=Stack
  Handler:  yaml/kind=Model
  Handler:  yaml/kind=Simulation
simulation.yaml
Stack
simulation.yaml
Model
simulation.yaml
Simulation

=== Report ===================================================================
Name: Channel 'expectedModelCount'
Path: /home/memgraph/.local/share/dse-graph/reports/static_validation.yaml
Version: 0.0.0
Date: 2025-05-27 10:15:29
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
+-------------+---------------+-------------+--------+
| CHANNELNAME | EXPECTEDCOUNT | ACTUALCOUNT | RESULT |
+-------------+---------------+-------------+--------+
| Network     |             1 |           1 | PASS   |
| physical    |             2 |           2 | PASS   |
+-------------+---------------+-------------+--------+
Evaluation: Report Passed
Query: Model to Channel Mapping
Cypher:
    MATCH (st:Stack)-[:Has]->(mi:ModelInst)-[a:Alias]->(ch:Channel)
    WITH mi, a, ch
    RETURN mi.name AS modelInstName, a.name as alias, ch.name AS channelName

Results:
+---------------+-----------------+-------------+
| MODELINSTNAME | ALIAS           | CHANNELNAME |
+---------------+-----------------+-------------+
| input         | scalar          | physical    |
| linear        | signal_channel  | physical    |
| linear        | network_channel | Network     |
+---------------+-----------------+-------------+


=== Report ===================================================================
Name: Count 'ModelInst' in AST and SIM
Path: /home/memgraph/.local/share/dse-graph/reports/static_validation.yaml
Version: 0.0.0
Date: 2025-05-27 10:15:29
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
+-------------------+-------------------+--------+
| ASTMODELINSTCOUNT | SIMMODELINSTCOUNT | RESULT |
+-------------------+-------------------+--------+
|                 1 |                 2 | FAIL   |
+-------------------+-------------------+--------+
Evaluation: Report Failed
Hint: The number of Model Instances in AST do not match the number of Model Instances in SIM.



=== Summary ===================================================================
[PASS] Channel 'expectedModelCount'
[FAIL] Count 'ModelInst' in AST and SIM
Ran 2 Reports | Passed: 1 | Failed: 1
```
</details>



### Duplicate Writes

Check a simulation configuration for Signals which are written by multiple
Models. In such instances the value of a signal becomes non-deterministic - such
simulations may be incorrectly configured.

<details>
<summary>Duplicate Writes - Simulation configuration check</summary>

```bash
...
Pinging Memgraph...
...
Running command: drop
...
$ dse-report --name="Duplicate Writes Check" examples/graph/duplicate_writes/sim_with_error
Running command: report
Options:
  db             : bolt://localhost:7687
  list           : false
  list-all       : false
  list-tags      : false
  name           :
  reports        :
  tag            :
2025/05/22 20:12:27 INFO Connect to graph db=bolt://localhost:7687
  Handler:  yaml/kind=Stack
  Handler:  yaml/kind=Model
simulation.yaml
Stack
simulation.yaml
Model
  Handler:  yaml/kind=Model
model.yaml
Model
  Handler:  yaml/kind=SignalGroup
signalgroup.yaml
SignalGroup
  Handler:  yaml/kind=Model
model.yaml
Model
  Handler:  yaml/kind=SignalGroup
  Handler:  yaml/kind=SignalGroup
signalgroup.yaml
SignalGroup
signalgroup.yaml
SignalGroup
2025/05/27 10:15:48 Relationship created: Channel(ID: 7421) -> Represents -> SignalGroup(ID: 7439)
2025/05/27 10:15:48 Skipping channel ID 7428 (labelCount: 1, selectorCount: 2)
2025/05/27 10:15:48 Relationship created: Channel(ID: 7428) -> Represents -> SignalGroup(ID: 7455)
2025/05/27 10:15:48 Relationship created: Channel(ID: 7421) -> Represents -> SignalGroup(ID: 7448)
2025/05/27 10:15:48 Skipping channel ID 7421 (labelCount: 1, selectorCount: 2)

=== Report ===================================================================
Name: Duplicate Writes Check
Path: /home/memgraph/.local/share/dse-graph/reports/duplicate_writes.yaml
Version: 0.0.0
Date: 2025-05-27 10:15:48
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
+-------------------------------------------------------------------------+----------------+
| SIMBUS_CHANNEL                                                          | COMMON_SIGNALS |
+-------------------------------------------------------------------------+----------------+
| {7418 7418 [Sim SimbusChannel] map[expectedModelCount:2 name:physical]} | [factor]       |
+-------------------------------------------------------------------------+----------------+
Evaluation: Report Failed


=== Summary ===================================================================
[FAIL] Duplicate Writes Check
Ran 1 Reports | Passed: 0 | Failed: 1
```
</details>



### Stack

Check that a simulation configuration contains only unique ModelInstance names
and that all Model UIDs are unique and non-zero.

<details>
<summary>Stack - Simulation configuration check</summary>

```bash
...
Pinging Memgraph...
...
Running command: drop
...
$ dse-report --name="ModelInstance Name Check;Model UID Check" examples/graph/stack/sim_with_error
Running command: report
Options:
  db             : bolt://localhost:7687
  list           : false
  list-all       : false
  list-tags      : false
  name           :
  reports        :
  tag            :
2025/05/22 20:18:33 INFO Connect to graph db=bolt://localhost:7687
  Handler:  yaml/kind=Stack
  Handler:  yaml/kind=Stack
  Handler:  yaml/kind=Stack
  Handler:  yaml/kind=Model
simulation.yaml
Stack
simulation.yaml
Stack
simulation.yaml
Stack
simulation.yaml
Model

=== Report ===================================================================
Name: ModelInstance Name Check
Path: /home/memgraph/.local/share/dse-graph/reports/stack.yaml
Version: 0.0.0
Date: 2025-05-27 10:16:31
Query: Unique ModelInstance Name
Cypher:
    MATCH (:Stack)-[:Has]->(mi:ModelInst)
    WHERE mi.name IS NOT NULL
    WITH mi.name AS name, count(*) AS count
    WHERE count > 1
    RETURN name, count

Results:
+-----------+-------+
| NAME      | COUNT |
+-----------+-------+
| counter_A |     2 |
+-----------+-------+
Evaluation: Report Failed


=== Report ===================================================================
Name: Model UID Check
Path: /home/memgraph/.local/share/dse-graph/reports/stack.yaml
Version: 0.0.0
Date: 2025-05-27 10:16:31
Query: Unique Non-zero Model UID
Cypher:
    MATCH (:Stack)-[:Has]->(mi:ModelInst)
    WHERE mi.uid IS NOT NULL
    WITH mi.uid AS uid, collect(mi.name) AS names, count(*) AS count
    WHERE uid = "0" OR count > 1
    RETURN uid, count, names

Results:
+-----+-------+-----------------------+
| UID | COUNT | NAMES                 |
+-----+-------+-----------------------+
| 43  |     2 | [counter_B counter_C] |
| 0   |     1 | [counter_A]           |
+-----+-------+-----------------------+
Evaluation: Report Failed


=== Summary ===================================================================
[FAIL] ModelInstance Name Check
[FAIL] Model UID Check
Ran 2 Reports | Passed: 0 | Failed: 2
```
</details>

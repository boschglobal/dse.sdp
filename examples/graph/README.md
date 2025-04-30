# Graph Report Examples

## Introduction

Use the Graph tool to validate simulations using a collection of Reports which perform analysis of simulation configuration files and simulation operational traces.


## Usage

```bash
# Start the memgraph container (using a make target).
$ make graph

# Clear the graph (of existing/previous content).
$ dse-graph drop -all

# Import simulation configuration files.
$ dse-graph import examples/graph/static_validation/sim_good

# Generate a report.
$ dse-graph report static_validation.yaml
...
=================== Summary ===================
Ran 2 Reports | Passed: 1 | Failed: 1
Failed Reports: Count 'ModelInst' in AST and SIM
===============================================
```


## Reports

### Static Validation

Performs a collection of static validation checks on simulation configuration files.

<details>
<summary>Static Validation - Simulation with no errors</summary>

```bash
$ dse-graph drop --all
$ dse-graph import examples/graph/static_validation/sim_good
$ dse-graph report static_validation.yaml
Running command: report
Options:
  db             : bolt://localhost:7687
  tag            : 
2025/04/30 08:51:27 INFO Connecting to graph db=bolt://localhost:7687

2025/04/30 08:51:27 INFO Report name: Channel 'expectedModelCount'
2025/04/30 08:51:27 INFO Path to Report: /home/codespace/.local/share/dse-graph/reports/static_validation.yaml
2025/04/30 08:51:27 INFO Query Name: Expected Count: 
MATCH (st:Stack)-[:Has]->(mi:ModelInst)-[:Alias]->(ch:Channel)
WITH ch.name AS channelName, COUNT(DISTINCT mi) AS actualCount, st
MATCH (st)-[:Has]->(simbus:Simbus)-[:Has]->(sc:SimbusChannel)
WHERE sc.name = channelName
RETURN channelName,
      sc.expectedModelCount AS expectedCount,
      actualCount,
      CASE WHEN sc.expectedModelCount = actualCount THEN "PASS" ELSE "FAIL" END AS result

+-------------+---------------+-------------+--------+
| CHANNELNAME | EXPECTEDCOUNT | ACTUALCOUNT | RESULT |
+-------------+---------------+-------------+--------+
| Network     |             1 |           1 | PASS   |
| physical    |             2 |           2 | PASS   |
+-------------+---------------+-------------+--------+

2025/04/30 08:51:27 INFO Report Passed

2025/04/30 08:51:27 INFO Query Name: Model to Channel Mapping: 
MATCH (st:Stack)-[:Has]->(mi:ModelInst)-[a:Alias]->(ch:Channel)
WITH mi, a, ch
RETURN mi.name AS modelInstName, a.name as alias, ch.name AS channelName

+---------------+-----------------+-------------+
| MODELINSTNAME | ALIAS           | CHANNELNAME |
+---------------+-----------------+-------------+
| input         | scalar          | physical    |
| linear        | signal_channel  | physical    |
| linear        | network_channel | Network     |
+---------------+-----------------+-------------+

====================================================================================================

====================================================================================================

2025/04/30 08:51:27 INFO Report name: Count 'ModelInst' in AST and SIM
2025/04/30 08:51:27 INFO Path to Report: /home/codespace/.local/share/dse-graph/reports/static_validation.yaml
2025/04/30 08:51:27 INFO Query Name: Expected Count: 
MATCH (fl:File)-[:Contains]->(st:Stack)-[:Has]->(mi:ModelInst)
WITH fl, COUNT(DISTINCT mi) AS countSim
MATCH (fl)-[:Contains]->(sim:Simulation)-[:Has]->(st2:Stack)-[:Has]->(mi2:ModelInst)
WITH countSim, COUNT(DISTINCT mi2) AS countAst
RETURN
    countAst AS astModelInstCount,
    countSim AS simModelInstCount,
    CASE WHEN countAst = countSim THEN "PASS" ELSE "FAIL" END AS result

+-------------------+-------------------+--------+
| ASTMODELINSTCOUNT | SIMMODELINSTCOUNT | RESULT |
+-------------------+-------------------+--------+
|                 2 |                 2 | PASS   |
+-------------------+-------------------+--------+

2025/04/30 08:51:27 INFO Report Passed

====================================================================================================

====================================================================================================

=================== Summary ===================
Ran 2 Reports | Passed: 2 | Failed: 0

===============================================
```
</details>


<details>
<summary>Static Validation - Simulation <b>with</b> errors</summary>

```bash
$ dse-graph drop --all
$ dse-graph import examples/graph/static_validation/sim_with_error
$ dse-graph report static_validation.yaml
Running command: report
Options:
  db             : bolt://localhost:7687
  tag            : 
2025/04/30 08:52:39 INFO Connecting to graph db=bolt://localhost:7687

2025/04/30 08:52:39 INFO Report name: Channel 'expectedModelCount'
2025/04/30 08:52:39 INFO Path to Report: /home/codespace/.local/share/dse-graph/reports/static_validation.yaml
2025/04/30 08:52:39 INFO Query Name: Expected Count: 
MATCH (st:Stack)-[:Has]->(mi:ModelInst)-[:Alias]->(ch:Channel)
WITH ch.name AS channelName, COUNT(DISTINCT mi) AS actualCount, st
MATCH (st)-[:Has]->(simbus:Simbus)-[:Has]->(sc:SimbusChannel)
WHERE sc.name = channelName
RETURN channelName,
      sc.expectedModelCount AS expectedCount,
      actualCount,
      CASE WHEN sc.expectedModelCount = actualCount THEN "PASS" ELSE "FAIL" END AS result

+-------------+---------------+-------------+--------+
| CHANNELNAME | EXPECTEDCOUNT | ACTUALCOUNT | RESULT |
+-------------+---------------+-------------+--------+
| Network     |             1 |           1 | PASS   |
| physical    |             2 |           2 | PASS   |
+-------------+---------------+-------------+--------+

2025/04/30 08:52:39 INFO Report Passed

2025/04/30 08:52:39 INFO Query Name: Model to Channel Mapping: 
MATCH (st:Stack)-[:Has]->(mi:ModelInst)-[a:Alias]->(ch:Channel)
WITH mi, a, ch
RETURN mi.name AS modelInstName, a.name as alias, ch.name AS channelName

+---------------+-----------------+-------------+
| MODELINSTNAME | ALIAS           | CHANNELNAME |
+---------------+-----------------+-------------+
| input         | scalar          | physical    |
| linear        | signal_channel  | physical    |
| linear        | network_channel | Network     |
+---------------+-----------------+-------------+

====================================================================================================

====================================================================================================

2025/04/30 08:52:39 INFO Report name: Count 'ModelInst' in AST and SIM
2025/04/30 08:52:39 INFO Path to Report: /home/codespace/.local/share/dse-graph/reports/static_validation.yaml
2025/04/30 08:52:39 INFO Query Name: Expected Count: 
MATCH (fl:File)-[:Contains]->(st:Stack)-[:Has]->(mi:ModelInst)
WITH fl, COUNT(DISTINCT mi) AS countSim
MATCH (fl)-[:Contains]->(sim:Simulation)-[:Has]->(st2:Stack)-[:Has]->(mi2:ModelInst)
WITH countSim, COUNT(DISTINCT mi2) AS countAst
RETURN
    countAst AS astModelInstCount,
    countSim AS simModelInstCount,
    CASE WHEN countAst = countSim THEN "PASS" ELSE "FAIL" END AS result

+-------------------+-------------------+--------+
| ASTMODELINSTCOUNT | SIMMODELINSTCOUNT | RESULT |
+-------------------+-------------------+--------+
|                 1 |                 2 | FAIL   |
+-------------------+-------------------+--------+

2025/04/30 08:52:39 INFO Hint !! The number of Model Instances in AST do not match the number of Model Instances in SIM.

2025/04/30 08:52:39 INFO Report Failed

=================== Summary ===================
Ran 2 Reports | Passed: 1 | Failed: 1
Failed Reports: Count 'ModelInst' in AST and SIM
===============================================
```
</details>



### Duplicate Writes Check

<details>
<summary>Duplicate Writes - Simulation configuration check</summary>

```bash
$ dse-graph drop --all
$ dse-graph import examples/graph/duplicate_writes/sim_with_error
$ dse-graph report duplicate_writes.yaml
...
```
</details>
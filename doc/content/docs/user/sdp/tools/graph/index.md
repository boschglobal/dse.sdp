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
$ graph import example/graph/sim
$ graph report
```


## Commands
The Graph tool includes the following commands and options:


### Drop

```bash
$ graph drop
```

#### Option All (-all)
```bash
$ graph drop --all
```

### Export

```bash
$ graph export <output-file>
```


### Import

```bash
$ graph import <input-dir>
```

### Report

```bash
$ graph report [--db=db_uri] <report-file>
```

#### Option Tag (-tag)

```bash
$ graph report [--tag=name --db=db_uri] <report-file>
```

### Ping

```bash
$ graph ping [--retry=count --db=db_uri]
```

## Container

The Graph Tool is also packaged as a Container and may be used outside of the
SDP.


```bash

```

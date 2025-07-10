---
title: "AST - AST Tools"
linkTitle: "AST"
weight: 100
tags:
- SDP
- CLI
github_repo: "https://github.com/boschglobal/dse.sdp"
github_subdir: "doc"
---


## Synopsis
AST Tools.

```bash
$ dse-ast <command> [flags]
```
The dse-ast toolchain provides commands for processing and transforming Abstract Syntax Trees (ASTs) in YAML format, based on input JSON.

## Commands
### convert  	
Transform the JSON into a YAML-based Abstract Syntax Tree (AST).

```bash
$ dse-ast convert -input <json_file_path> -output <yaml_ast_output_path>
```

### resolve
Resolve internal references within the AST to produce a fully linked version.

```bash
$ dse-ast resolve -input <yaml_ast_path> -output <yaml_ast_output_path>
```

### generate
Generate the final output simulation files based on the resolved AST.

```bash
$ dse-ast generate -input <yaml_ast_path> -output <output_path>
```
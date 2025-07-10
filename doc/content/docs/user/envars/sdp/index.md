---
title: "SDP Environment Variables"
linkTitle: "SDP"
weight: 600
tags:
- Envar
github_repo: "https://github.com/boschglobal/dse.sdp"
github_subdir: "doc"
---

## Container Specific Environment Variables

### Tool Specific

| Variable           | Default |
| ------------------ | ------- |
| <var>SIMER_IMAGE</var>      | `ghcr.io/boschglobal/dse-simer:latest` |
| <var>BUILDER_IMAGE</var>    | `ghcr.io/boschglobal/dse-builder:latest` |
| <var>REPORT_IMAGE</var>     | `ghcr.io/boschglobal/dse-report:latest` |


### Workflow Specific

| Variable                  | Default |
| ------------------------- | ------- |
| <var>FMI_IMAGE</var>      | `ghcr.io/boschglobal/dse-fmi` |
| <var>FMI_TAG</var>        | `latest` |

> Note: this pattern of <var>IMAGE</var> and <var>TAG</var> is repeated for all
> repos which contain workflows.

---
title: "SDP Developer Documentation"
linkTitle: "SDP"
cascade:
  - type: "docs"
weight: 20
---

Simulation Development Platform - Developer Documentation



## Debug Workflows

Build the workflow container.

~/git/working/dse.fmi$ make build fmi tools


Set env to pickup the local container

 export FMI_IMAGE=fmi
 2009  export FMI_TAG=test
 2010  export TASK_X_REMOTE_TASKFILES=1

Build the taskfile

make build
or
dse-parse2ast runnable.dse runnable.json
dse-ast convert -input runnable.json -output runnable.yaml
dse-ast resolve -input runnable.yaml
dse-ast generate -input runnable.yaml -output .

Edit the Taskfile uses to point to local versoin

dse.fmi-v1.1.31:
    taskfile: ../../..//dse.fmi/Taskfile.yml


Run task

task -y -v -f

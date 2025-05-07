---
title: "SDP Workflows"
linkTitle: "Workflows"
weight: 20
---

The SDP builds simulations using Git repositories which contain both models and
their associated workflows. The workflows are defined using
[Task](https://taskfile.dev/), and implemented with containerized tool-chains.
Additional metadata, embedded in the Taskfile, makes is possible for the SDP to
automate the task of configuring and operating the workflows.


## Layout

### Go based Tools

Models are typically written in the C Language using an established repository
layout. Within _that_ layout, workflow tools can be written in any programming
language. For Go based workflow tools, the following layout is suggested:

```text
extra
└── tools
    └── <tool>
        ├── build/package
        │   ├── Dockerfile
        │   └── entrypoint.sh   <-- entry point for container image
        ├── cmd/<tool>
        │   ├── main.go         <-- CLI configuration of commands
        │   └── main_test.go    <-- testscript interface (optional)
        ├── internal/app/       <-- command implementation and tests
        ├── pkg/                <-- reusable packages (optional)
        ├── test/testdata/      <-- testscript (txtar) based tests
        ├── vendor/             <-- vendored packages (optional)
        ├── Makefile
        ├── README.md
        ├── go.mod
        └── go.sum
tests
└── testscript
    └── e2e/        <-- repo level end-to-end tests
Makefile            <-- repo level Makefile
Taskfile            <-- definition of workflows and related metadata
```

> __Hint:__ Typically the tool name and repo name are identical.


## Containers

The general preference is to have one workflow/tool container per repository
which contains all necessary workflow tools. An additional `entrypoint.sh`
script can be added to the container to facilitate command selection/execution.


## Taskfile

Workflow automation is supported by Task which uses a Taskfile to define
workflows, and those workflows use the containerized tools exclusively. A
Taskfile will have features as described in the following sections, and working
examples can be found at the root of most DSE repos.


### Repo Metadata

Repo Metadata describes the following items:

* __Packaging__ - Download URL for compiled models belonging to a repo.
* __Container__ - The name of the repository containing images for the tool
  container.
* __Models__ - a list of the models, and associated data for each model, that is
  included in the package.
  * __Name__ - The name of the model.
  * __Display Name__ - How the name of the model should be displayed in the
    SDP/DSL.
  * __Path__ - The relative path of the model in the package archive.
  * __Mcl__ - Indicates that this is an MCL model.
  * __Workflows__ - lists workflows that may be used with this model.
  * __Platforms__ - lists platforms this model supports.
  * __Channels__ - lists fixed/pre-configured channels of the model.


<details>
<summary>Example Taskfile with Repo Metadata</summary>

```yaml
---
version: '3'

metadata:
  package:
    download: '{{.REPO}}/releases/download/v{{.TAG}}/Fmi-{{.TAG}}-{{.PLATFORM_ARCH}}.zip'
  container:
    repository: ghcr.io/boschglobal/dse-fmi
  models:
    dse.fmi.mcl:
      name: fmimcl
      displayName: dse.fmi.mcl
      path: fmimcl
      mcl: true
      workflows:
        - generate-fmimcl
        - generate-fmimodelc
        - generate-fmigateway
        - patch-signalgroup
      platforms:
        - linux-amd64
        - linux-x86
        - linux-i386
        - windows-x64
        - windows-x86
      channels:
        - alias: signal_channel
        - alias: network_channel
```

</details>


### Taskfile Global Vars

The Taskfile global vars should include the following items:

* __ENTRYDIR__ - Necessary for operation of end-to-end tests.
* __IMAGE__ - Specify the default location of the tool container repository.
* __TAG__ - Specify the container image tag selection.


> __Hint:__ Typically the tool name and repo name are identical.

<details>
<summary>Example Taskfile with Global Vars </summary>

```yaml
---
version: '3'

vars:
  # When running from E2E tests (i.e. Docker in Docker), the ENTRYDIR (for
  # Docker commands) must be set to the host relative path.
  ENTRYDIR: '{{if .SIM}}{{.ENTRYWORKDIR}}/{{.SIM}}{{else}}{{.PWD}}{{end}}'
  # Container image specification.
  FMI_IMAGE: '{{.FMI_IMAGE | default "ghcr.io/boschglobal/dse-fmi"}}'
  FMI_TAG: '{{if .FMI_TAG}}{{.FMI_TAG}}{{else}}{{if .IMAGE_TAG}}{{.IMAGE_TAG}}{{else}}latest{{end}}{{end}}'
```

</details>


### Workflow Task Definition

Workflow Tasks are the interface used by the SDP and represent an API interface.
They may call other tasks, as required, to implement a workflow. Workflow Tasks
also include a Metadata object which is used by the SDP/DSL to assist in
specifying variables when calling a Workflow Task.

The Metadata can include the following items:

* __Default__ - Indicate a default value for this variable (optional).
* __Hint__ - A short description of the variable. For complex variables also
  consider including an example in the hint.
* __Required__ - Indicates that this variable is required.


<details>
<summary>Example Taskfile with a Task Definition</summary>

```yaml

tasks:
  patch-signalgroup:
    desc: Patch changes into a generated Signal Group.
    run: always
    dir: '{{.USER_WORKING_DIR}}'
    label: dse:fmi:patch-signalgroup
    vars:
      # INPUT: '{{.INPUT}}'
      # PATCH: '{{.PATCH}}'
      REMOVE_UNKNOWN: '{{if .REMOVE_UNKNOWN}}--remove-unknown{{else}}{{end}}'
    cmds:
      - docker run --rm
          -v {{.ENTRYDIR}}:/sim
          {{.FMI_IMAGE}}:{{.FMI_TAG}}
          patch-signalgroup
            --input {{.INPUT}}
            --patch {{.PATCH}}
            {{.REMOVE_UNKNOWN}}
    requires:
      vars: [INPUT, PATCH]
    metadata:
      vars:
        INPUT:
          required: true
          hint: Path identifying the Signal Group to be patched.
        PATCH:
          required: true
          hint: URI identifying the patch file to use.
        REMOVE_UNKNOWN:
          required: false
          hint: Remove unknown items (i.e. not in the patch file).
          default: false
```

</details>

> __Note:__ The commented `vars`, necessary for the `requires:vars` to work, are
> retained for clarity/documentation.


## Developer Notes

### Vendoring

Workflow tools that make use of private Go packages may need to be vendored
(especially if the Workflows tools are in public repos).


### Testing

Workflow tools should be unit tested. In some cases Testscript/Txtar tests may
be a good alternative for writing tests. A repo should also include end-to-end
tests where the focus in on the operation of the Models and Workflows together.

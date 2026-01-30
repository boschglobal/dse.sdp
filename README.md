# Dynamic Simulation Environment - Simulation Development Platform

[![Containers](https://github.com/boschglobal/dse.sdp/actions/workflows/containers.yaml/badge.svg)](https://github.com/boschglobal/dse.sdp/actions/workflows/containers.yaml)
[![Super Linter](https://github.com/boschglobal/dse.sdp/actions/workflows/super-linter.yaml/badge.svg)](https://github.com/boschglobal/dse.sdp/actions/workflows/super-linter.yaml)
<br/>
[![Open in GitHub Codespaces](https://github.com/codespaces/badge.svg)](https://codespaces.new/boschglobal/dse.sdp?quickstart=1)


## Introduction

Simulation Development Platform for the Dynamic Simulation Environment (DSE) Core Platform.

### User Guide

* [SDP Setup Guide][ug_sdp]
* [DSE Script and Simulation Builder][ug_builder]
* [Simer - Simulation Runner][ug_simer]
* [Lua Models][ug_lua]
* [SDP Environment Variables][ug_env]


### Project Structure

```text
dse.sdp
└── .devcontainer/          <-- Devcontainer definition.
    └── Dockerfile-builder  <-- Builder tool container appliance (Dockerfile).
└── actions                 <-- GitHub actions (for Builder, Report and Simer containers).
└── ast                     <-- AST tools.
└── doc                     <-- Project documentation.
└── dsl                     <-- DSL parser (using Chevrotain).
└── examples
    └── graph               <-- Graph examples (for Report tool)
    └── models              <-- Model examples, used by E2E tests.
    └── openloop            <-- Open Loop simulation using FMU based Linear Equation model.
    └── openloop_lua        <-- Open Loop simulation implemented with Lua models.
    └── notebook            <-- Jupyter base simulation example.
    └── vscode              <-- VS Code integration examples.
└── graph                   <-- Graph tool (including Report tool).
    └── build/package/
        └── report/         <-- Report tool container appliance (Dockerfile).
    └── cmd/graph/
        └── reports/        <-- Reports, packaged with Report container.
└── licenses                <-- Third Party Licenses.
└── lsp                     <-- VS Code Language Server.
└── tests
    └── e2e                 <-- End-to-end (E2E) tests.
    └── scripts             <-- Scripts used by E2E tests.
    └── testdata            <-- Testdata, used by E2E tests.
└── Makefile                <-- Repo level Makefile.
└── Taskfile.yml            <-- Taskfile containing supporting automation for E2E tests.
```


## Usage

> Hint: Codespaces is known to work with Chrome and Edge browsers. Firefox may prevention operation via Firefox's Setting "Enhanced Tracking Protection" (try setting to Standard to resolve the issue).


### Examples

* [VS Code Integration Example](examples/vscode/README.md)
* [Notebook Based Simulation Example](examples/notebook/README.md)
* [Open Loop Simulation Example](examples/openloop/README.md)
* [Open Loop Simulation with Lua models](examples/openloop_lua/README.md)


### Running ModelC Example Simulations

Start a Codespace, then type the following commands in the terminal window.

```bash
# Check your environment.
$ dse-env
DSE_MODELC_VERSION=2.1.30
DSE_REPORT_IMAGE=ghcr.io/boschglobal/dse-report:latest
DSE_SIMER_IMAGE=ghcr.io/boschglobal/dse-simer:latest

# Setup the examples (will download ModelC examples).
$ make examples
$ ls out/examples/modelc/
benchmark/  binary/  extended/  gateway/  gdb/  minimal/  ncodec/  runtime/  simer/  transform/

# Validate a simulation using the Report tool.
$ dse-report out/examples/modelc/minimal
...
=== Summary ===================================================================
[PASS] Duplicate Writes Check
[PASS] ModelInstance Name Check
[PASS] Model UID Check
[PASS] Channel 'expectedModelCount'
[PASS] Count 'ModelInst' in AST and SIM
Ran 5 Reports | Passed: 5 | Failed: 0

# Run a simulation using Simer.
$ dse-simer out/examples/modelc/minimal
...
2) Run the Simulation ...
2) Controller exit ...


# Build and run a simulation using the DSL.
$ cd examples/vscode
$ make build
...
$ make report
...
$ make run
...
```

> Hint: Find more information about the Simer [command options here](https://boschglobal.github.io/dse.doc/docs/user/simer/#options).

> Hint: Find more information about the Report [command options here](https://boschglobal.github.io/dse.doc/docs/user/report/#options).



## Build

### WSL Linux (local development)

```bash
# Install NVM, and node, if necessary.
$ curl -o- https://raw.githubusercontent.com/nvm-sh/nvm/v0.40.3/install.sh | bash
$ nvm ls-remote
$ nvm install v22.15.0

# Install prerequisites.
$ sudo npm install -g vsce
$ sudo npm install -g http-server
$ sudo npm install -g typescript

# Set your path to include ~/.local/bin if necessary. Permanent alterations
# can be made to your '~/.bashrc' or '~/.profile' file.
$ export PATH="$HOME/.local/bin:$PATH"

# Clone the repo.
$ git clone https://github.boschdevcloud.com/fsil/dse.sdp.git
$ cd dse.sdp

# Setup the SDP (local install).
$ make
$ make build install

# Build containerised tools (container/docker install).
$ make docker

# Optionally use local container images.
$ DSE_BUILDER_IMAGE=dse-builder:test
$ DSE_REPORT_IMAGE=dse-report:test
$ DSE_SIMER_IMAGE=dse-simer:test

# Run tests.
$ make run_graph
$ make test
$ make test_e2e

# Generate documentation.
$ make generate

# Remove (clean) temporary build artifacts.
$ make clean
$ make cleanall
```


## Developer Notes

### Dev Containers Extension for VS Code and WSL

> Note: Dev Containers is not supported by VS Codium (workarounds may still exist).

#### Install VS Code Extensions

1. Install the Dev Containers extension (Ctrl-Shift-X, then search "Dev Containers").

2. Install the WSL extension (Ctrl-Shift-X, then search "WSL").

3. Install the Codespaces extension (Ctrl-Shift-X, then search "Codespace").

4. From the Remote Explorer, select a WSL Target and then click `Connect in New Window`. A new VS Code editor will open.

5. Press `F1` to bring up the Command Palette and type `Dev Containers reopen`. You may be prompted to install docker into WSL. Even if already installed, proceed to install Docker again (in your WSL)


WSL indicator shows in the bottom left corner of the VS Code window.
The Remote Explorer, added by the WSL extension, will show available WSL Targets.


#### Start VS Code with a WSL Target (and DevContainer)

**Using WSL Bash Terminal:**

```bash
# Open a new VS Code editor connected to this repo.
cd ~/git/workspace/dse.sdp
code.

# In VS Code, press `F1` and type `Dev Containers reopen`.

# Open a bash terminal to access the DevContainer (i.e. local Codespace).
codespace ➜ /workspaces/dse.sdp (main) $ which task
/usr/local/bin/task
codespace ➜ /workspaces/dse.sdp (main) $
```

**Using VS Code:**

1. Click (bottom left) `Open a Remote Window`, then command `Connect to WSL`.

2. From the _Remote Explorer_, select a _WSL Target_ and then click `Connect in Current Window`.<br/>
  A new VS Code editor will open.

3. Select `Open Folder` and naviage to your dse.sdp repo. Previouly opened repos will be listed in the _Remote Explorer_.


After that, the container will build ... and eventually you will have the Codespace avaiable in your Terminal Window.


### Proxy Setup when running _inside_ a DevContainer

[proxy setup](https://docs.docker.com/engine/cli/proxy/)

~/git/working/dse.sdp$ cat ~/.docker/config.json



## Additional Resources

* [Chevrotain](https://chevrotain.io/docs/) Parser
* [Dev Containers](https://marketplace.visualstudio.com/items?itemName=ms-vscode-remote.remote-containers) for VS Code
* [Simer](https://boschglobal.github.io/dse.doc/docs/user/simer/) (and [source code](https://github.com/boschglobal/dse.modelc/tree/main/extra/tools/simer))



## Contribute

Please refer to the [CONTRIBUTING.md](./CONTRIBUTING.md) file.



## License

Dynamic Simulation Environment FMI Library is open-sourced under the Apache-2.0 license.
See the [LICENSE](LICENSE) and [NOTICE](./NOTICE) files for details.


### Third Party Licenses

[Third Party Licenses](licenses/)



<!---  Links --->
[ug_builder]: https://boschglobal.github.io/dse.doc/docs/user/builder/
[ug_simer]: https://boschglobal.github.io/dse.doc/docs/user/simer/
[ug_lua]: https://boschglobal.github.io/dse.doc/docs/user/models/lua/
[ug_sdp]: https://boschglobal.github.io/dse.doc/docs/user/guides/sdp/
[ug_env]: https://boschglobal.github.io/dse.doc/docs/user/envars/sdp/

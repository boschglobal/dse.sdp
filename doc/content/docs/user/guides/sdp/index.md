---
title: "Guide : Simulation Development Platform (SDP)"
linkTitle: "SDP"
weight: 50
tags:
- SDP
github_repo: "https://github.com/boschglobal/dse.sdp"
github_subdir: "doc"
---

## Synopsis

Simulation Development Platform (SDP) using Codespaces or DevContainers to code, build, and run DSE Simulations.



## Setup

### GitHub Codespaces

GitHub Codespaces provides a cloud-hosted development environment which is
pre-configured for the SDP. This environment includes SDP Extensions and can
be immediately used to build and run simulations.


#### Steps to use Codespaces:

1. Go to the repository on GitHub.
2. Click the <b>Code</b> button and choose <b>Open with Codespaces</b>.
3. If you don’t have an existing codespace, click <b>New codespace</b> to create one.
4. After some moments your Codespace will be ready.


### Dev Containers

VS Code Dev Containers can be configured to run the SDP within Visual Studio,
and the SDP Extensions can also be installed. This approach requires a Docker
environment (e.g. WSL2 or Docker Desktop).


#### Steps to use a Dev Container:

> TODO: This section needs to be updated.

1. Configure the Dev Container.

2. Install the extension using VS Code GUI:
	1. Open the Extensions view by pressing `Ctrl+Shift+X`.
	2. Click the `...` (More Actions) menu in the top-right corner of the Extensions panel.
	3. Select `Install from VSIX...`.
	4. Navigate to the `lsp/out/bin` folder and select the generated `dse.vsix` file.


### Native Linux

The SDP Builder and Report tool, as well as the Simer simulation run-time, are
all containerized tools which can be configured and used in a Linux environment.


## Usage

### GitHub Workflows

> TODO: This section needs to be updated.


### VS Code Extension

#### VS Code DSE Commands

The following commands are available via the Command Palette (`Ctrl+Shift+P`) when the SDP extension is installed:

| Command | Description |
|---------|-------------|
| Build (`DSE: Build`) | Generates `simulation.yaml` and `Taskfile.yml` from the active `.dse` file. This prepares the simulation environment. |
| Check (`DSE: Check`) | Analyzes the simulation graph and produces a report to help visualize and verify the structure of the simulation. |
| Run (`DSE: Run`) | Executes the simulation using the currently configured simulation definition. |
| Clean (`DSE: Clean`) | Performs a clean operation using `task: clean`, removing generated artifacts and build files. |
| Cleanall (`DSE: Cleanall`) | Performs a deep clean via `task: cleanall`, removing all outputs and intermediate data. |


#### Live AST View

The extension supports live viewing of the models and channels derived from `.dse` files

##### To view the AST preview

1. Open a supported `.dse` file in the VS Code editor.
2. Click the `Open Preview` button in the upper-right corner of the editor window.

###### Alternatively, you can use keyboard shortcuts

- Press **`Ctrl + K V`** to open preview in a side panel.
- Press **`Ctrl + Shift + V`** to open preview in the main panel.

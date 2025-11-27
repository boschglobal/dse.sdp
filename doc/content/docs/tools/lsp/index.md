---
title: "LSP - LSP Tools"
linkTitle: "LSP"
weight: 100
tags:
- SDP
- CLI
github_repo: "https://github.com/boschglobal/dse.sdp"
github_subdir: "doc"
---


## Synopsis
LSP Tools.

## Live AST View
The extension supports live viewing of the models and channels derived from `.dse` files

##### To view the AST preview

1. Open a supported `.dse` file in the VS Code editor.
2. Click the `Open Preview` button in the upper-right corner of the editor window.

###### Alternatively, you can use keyboard shortcuts

- Press **`Ctrl + K V`** to open preview in a side panel.
- Press **`Ctrl + Shift + V`** to open preview in the main panel.


#### Important: One-Time Setup for Local Environments

If you're working locally in VS Code (not in GitHub Codespaces), you may need to manually install required global dependencies the first time:

```powershell
cd dse.sdp/dsl
npm install -g --force
```
This will ensure that all necessary global dependencies are installed properly.



## Commands
The following commands are available via the Command Palette (`Ctrl+Shift+P`) when the SDP extension is installed:

| Command | Description |
|---------|-------------|
| Build (`DSE: Build`) | Generates `simulation.yaml` and `Taskfile.yml` from the active `.dse` file. This prepares the simulation environment. |
| Check (`DSE: Check`) | Analyzes the simulation graph and produces a report to help visualize and verify the structure of the simulation. |
| Run (`DSE: Run`) | Executes the simulation using the currently configured simulation definition. |
| Clean (`DSE: Clean`) | Performs a clean operation using `task: clean`, removing generated artifacts and build files. |
| Cleanall (`DSE: Cleanall`) | Performs a deep clean via `task: cleanall`, removing all outputs and intermediate data. |

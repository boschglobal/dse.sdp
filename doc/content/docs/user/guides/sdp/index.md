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

## Labels & Selectors

### Signal Group Selection

A `SignalGroup` is not loaded by a model directly by _name_. Instead, each Model _channel_ defines a set of `selectors` and the runtime searches all loaded `SignalGroup` documents for those whose `metadata.labels` match those selectors. This indirection allows a `SignalGroup` to be assembled from several documents, and allows a model to be reused against different signal definitions without changing the model code.

When a simulation is run (e.g. via Simer), the flow is:

1. The runtime iterates the channels defined on the Model (or the channel of a Model Instance in the Stack).
2. For each channel, a _selector_ is built from that channel's `selectors` mapping (`spec/channels[].selectors`). A selector on the Model Instance (in the Stack) takes priority over the selector on the Model definition.
3. Every loaded `SignalGroup` is inspected. A `SignalGroup` matches only when **all** selector key/value pairs are present (with identical values) in the `SignalGroup`'s `metadata.labels`. Matching is a logical **AND**; any missing or mismatched label excludes the `SignalGroup`.
4. The signals of _all_ matching `SignalGroup` documents are combined into the Signal Vector for that channel.

> Note: A `SignalGroup` may declare **additional** labels beyond those named in the channel selectors — extra labels are simply ignored. Only the labels named by the selectors are required.


### Which labels are required?

The labels you must define on a `SignalGroup` are exactly the keys listed in the matching channel `selectors` — no more, no less:

- If a channel selector is `channel: <value>` only, then only the `channel` label is required. A `model` label (or any other) is optional and ignored.
- If a channel selector lists both `channel:` and `model:`, then **both** the `channel` and `model` labels are required on the `SignalGroup`, and both values must match, otherwise the signals will not be loaded.

> Note: The SDP **generated** `simulation.yaml` (the `Stack`) defines selectors on each _Model Instance_ channel (`spec/models[].channels[].selectors`) and **by default includes both `channel` and `model`**. Because a Model-Instance selector takes priority over the Model definition, **both the `channel` and `model` labels are required** on every `SignalGroup` in such a simulation.

**Generated Stack (simulation.yaml) — Model Instance channel selectors :**
```yaml
kind: Stack
metadata:
  name: models
spec:
  models:
    - name: can_public
      model:
        name: dse.network
      channels:
        - alias: signal_channel
          name: com_phys
          selectors:
            channel: signal_vector    # Selector -> requires matching label.
            model: can_public         # Selector -> requires matching label.
        - alias: network_channel
          name: Network
          selectors:
            channel: network_vector
            model: can_public
```

**Matching Scalar Signal Group :**
```yaml
kind: SignalGroup
metadata:
  name: can_public
  labels:
    channel: signal_vector    # Matches the 'signal_channel' selector.
    model: can_public
spec:
  signals:
    - signal: LatteralAccel
```

**Matching Binary Signal Group :**
```yaml
kind: SignalGroup
metadata:
  name: can_public
  labels:
    channel: network_vector   # Matches the 'network_channel' selector.
    model: can_public
  annotations:
    vector_type: binary
spec:
  signals:
    - signal: CAN_BUS_PUBLIC
      annotations:
        mime_type: application/x-automotive-bus;interface=stream;type=frame;bus=can;schema=fbs;bus_id=4;node_id=11;interface_id=1
        network: can_public
```

In the example above, because each channel selector lists both `channel` and `model`, **both** labels are mandatory on the corresponding `SignalGroup`. Omitting the `channel` label (or the `model` label) would cause the selector match to fail and the model would start with an empty Signal Vector for that channel.

# E2E DSE SDP Toolchain Example

## Introduction

This example introduces the DSE SDP Toolchains which can be used to build a run a simulation written in a
simple DSE Simulation Language. The DSE Simulation Language is supported by the DSE SDP integration with
VS Code, including a Simulation visualization.


### Example DSE Simulation Language

```text
simulation arch=linux-amd64
channel physical

uses
dse.modelc https://github.com/boschglobal/dse.modelc v2.1.23
dse.fmi https://github.com/boschglobal/dse.fmi v1.1.23
linear_fmu https://github.com/boschglobal/dse.fmi/releases/download/v1.1.23/Fmi-1.1.23-linux-amd64.zip path=examples/fmu/linear/fmi2/linear.fmu

model input dse.modelc.csv
channel physical signal_channel
envar CSV_FILE model/input/data/input.csv
file input.csv input/openloop.csv
file signalgroup.yaml input/signalgroup.yaml

model linear dse.fmi.mcl
channel physical signal_channel
workflow generate-fmimcl
var FMU_DIR uses linear_fmu
var OUT_DIR {{.PATH}}/data
var MCL_PATH {{.PATH}}/lib/libfmimcl.so
```


### Simulation Project Layout

```text
(Git committed files)
L- openloop.dse        DSE Simulation Language.
L- extra           
  L- openloop.yaml     Reference ASL for the openloop simulation.
L- input
  L- openloop.csv      Data for the input model.
  L- signalgroup.yaml  Supporting SignalGroup for openloop.csv.

(Generated files)
L- simulation.yaml     Generated simulation (contains stacks).
L- Taskfile.yaml       Generated taskfile (constructs the simulation).
L- out     
  L- cache       	   Cached metadata.
  L- downloads         Downloaded content (models, tools ...).
  L- sim               Generated simulation (with Simer layout).
```


## Usage

### Using VS Code



### Using the CLI

Operating the E2E example via the CLI introduces the complete toolchain.


#### Toolchain Setup

> Note: the Codespace will have all Toolchains already installed and ready to use.

```bash
# Build and install the toolchains.
$ make
$ make install

# Setup additional commands.
$ function dse-simer() { ( if test -d "$1"; then cd "$1" && shift; fi && docker run -it --rm -v $(pwd):/sim -p 2159:2159 -p 6379:6379 $DSE_SIMER_IMAGE "$@"; ); }
$ export -f dse-simer

# Configure Task.
$ export TASK_X_REMOTE_TASKFILES=1
```


#### Build and Run the Project

```bash
$ cd examples/e2e

# Compile the DSL and generate the Simulation AST.
$ dse-parse2ast openloop.dse openloop.json
$ dse-ast convert -input openloop.json -output openloop.yaml
$ dse-ast resolve -input openloop.yaml

# Build the Simulation.
$ dse-ast generate -input openloop.yaml -output .
$ task -y -v

# Run the Simulation.
$ dse-simer out/sim
...

```

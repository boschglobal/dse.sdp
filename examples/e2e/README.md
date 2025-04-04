# VS Code Integration Example

## Introduction

This example demonstrates the VS Code Integration including: DSE Support, Preview, Commands and Measurement Plotting.


## Usage

### Using VS Code Integration with a Codespace

1. Open a Codespace for the [DSE SDP](https://github.com/boschglobal/dse.sdp) repository.
2. Navigate to the examples/vscode folder.
3. Open the file `openloop.dse` in an editor.
4. Use the preview button (or type `Ctrl + K then V`) to view a visualization of the Simulation.
5. Build the simulation: type `Ctrl + Shift + P` and then select command `DSE:Build`.
6. Run the simulation: type `Ctrl + Shift + P` and then select command `DSE:Run`.
7. Open the generated measurement file `out/sim/measurement.csv`.


### Using Terminal

```bash
$ cd examples/vscode

# Build the simulation.
$ make build

# Run the simulation.
$ make run

# Cleanup any generated files.
$ make clean
```


## Details

### Simulation Project Layout

```text
(Git committed files)
L- .gitignore          Git ignore file.
L- Makefile            Makefile automation.
L- openloop.dse        Simulation script written in DSE Simulation Language.
L- post_run.sh         Post simulation-run script (processes measurement).
L- README.md           Readme file supporting this example.        
L- extra           
  L- openloop.yaml     Reference ASL for the openloop simulation.
L- input
  L- openloop.csv      Data for the input model.
  L- signalgroup.yaml  Supporting SignalGroup for openloop.csv.

(Generated files)
L- simulation.yaml     Generated simulation (contains stacks).
L- Taskfile.yaml       Generated taskfile (constructs the simulation).
L- out     
  L- cache       	     Cached metadata.
  L- downloads         Downloaded content (models, tools ...).
  L- sim               Generated simulation (with Simer layout).
    L- measurement.mf4 Measurement file (from the linear model).
    L- measurement.csv Measurement file converted to CSV format.
```

# FMU Gateway Example

> Note: This example runs on Windows using the Simer Runtime.

## Introduction

This example demonstrates how a simulation can be packaged into an FMU Gateway.


## Example Layout

```text
examples
└── fmugw
    └── data
        └── input.csv                 <-- Input data (used by input/csv model).
        └── input.yaml                <-- Signal Group definition for input data.
    └── Makefile                      <-- Automation targets: build, report, run.
    └── simulation.dse                <-- Simulation definition written in DSE Script.
    └── README.md                     <-- Usage instructions for the example.
```


## Usage

```bash
# Help
$ make

# Build the simulation (Linux only).
$ make build

# Run the simulation.
$ make run

# Delete generated simulation artefacts.
$ make clean
# Delete cached downloads.
$ make cleanall
```

### Operation

> Note: This example runs the simulation on Windows (Powershell).

```ps

```

### Simulation Layout

```text
examples/fmugw/out
└── cache                               <-- Cache files.
└── download                            <-- Download cache.
└── sim
    └── bin
        └── redis-server.exe            <-- Redis server.
        └── simer.exe                   <-- Simer Runtime executable.
        └── x64
            └── modelc.exe              <-- ModelC Runtime executable.
            └── simbus.exe              <-- SimBus executable.
        └── x86
            └── modelc.exe              <-- ModelC Runtime executable (32bit).
            └── simbus.exe              <-- SimBus executable (32bit).
    └── data
        └── simulation.yaml             <-- Simulation definition, contains stacks.
    └── model/input
        └── data
            └── input.csv               <-- Scenario definition in CSV format.
            └── model.yaml              <-- Model definition.
            └── signalgroup.yaml        <-- Generated SignalGroup (based on input.csv).
        └── lib
            └── libcsv.dll              <-- Model implementation (shared library).
    └── model/linear
        └── data
            └── model.yaml              <-- Generated model definition.
            └── signalgroup.yaml        <-- Generated SignalGroup.
        └── lib
            └── libfmimcl.dll           <-- FMI MCL.
        └── linear_fmu                  <-- Downloaded FMU.
            └── modelDescription.xml
            └── resources
            └── binaries/win64
                └── fmu2linear.dll      <-- FMU implementation (shared library).
```

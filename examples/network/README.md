# Gateway Example

## Introduction

This example demonstrates building a Model/Simulation that support binary/mime-typed signals within the DSL.


## Usage

### WSL Linux

```bash
# Clone the repo.
$ git clone https://github.boschdevcloud.com/fsil/dse.sdp.git
$ cd dse.sdp

# Network model setup
$ cd examples/models
$ make

# Build and run the example simulation.
$ cd examples/network
$ make
$ make simer

# Delete the simulation and generated files
$ make clean
```


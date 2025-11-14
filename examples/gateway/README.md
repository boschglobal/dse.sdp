# Gateway Example

## Introduction

This example demonstrates building a Gateway Model/Simulation which support the concept of an external model in the DSL.


## Usage

### WSL Linux

```bash
# Clone the repo.
$ git clone https://github.boschdevcloud.com/fsil/dse.sdp.git
$ cd dse.sdp

# Gateway setup
$ cd examples/models
$ make

# Build and run the example simulation.
$ cd examples/gateway
$ make
$ make simer
$ make gateway

# Delete the simulation and generated files
$ make clean
```


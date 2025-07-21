# Runnable FMU Example

## Introduction

This example demonstrates building a Runnable Model/Simulation which is packaged
as an FMU using the DSE FMI ModelC FMU.


## Usage

### Authentication Tokens

This example requires the following authentication tokens to be set in the
operating environment:

* For downloading artefacts and container images from BDC Artifactory:
  * AR_USER (`abc1abt@bosch.com`)
  * AR_TOKEN (xxxx)
* Accessing BDC (i.e. FSIL) github repositories:
  * GHE_USER (abc1abt)
  * GHE_TOKEN (zzzzz)
  * GHE_PAT (ghp_yyyyy)

### Codespace / Devcontainer

```bash
TODO
```


### WSL Linux

```bash
# Clone the repo.
$ git clone https://github.boschdevcloud.com/fsil/dse.sdp.git
$ cd dse.sdp

# Build and run the example simulation.
$ cd examples/runnable
$ make build
$ make task

# Check the simulation.
$ make report

# Run the simulation using Simer (note: does not run the packaged FMU)
$ make simer
$ make simer-modelc-runtime

# Run the simulation using FMPy
# TODO update with python script.

# Delete the simulation and generated files
$ make clean
# Delete all downloaded files (in addition to clean).
$ make cleanall
```

# Runnable FMU Example

## Introduction

This example demonstrates building a Runnable Model/Simulation which is packaged
as an FMU using the DSE FMI ModelC FMU.


## Usage

### Authentication Tokens

This example requires the following authentication tokens to be set in the
operating environment:

* For downloading artefacts and container images from BDC Artifactory:
    * AR_USER (abc1abt@bosch.com)
    * AR_TOKEN (xxxx)
* Accessing BDC (i.e. FSIL) github repositories:
    * GHE_USER (abc1abt)
    * GHE_TOKEN (zzzzz)
    * GHE_PAT (ghp_yyyyy)


### VS Code with Devcontainer (WSL Backend)


### WSL Linux

```bash
# Install NVM, and node, if necessary.
$ curl -o- https://raw.githubusercontent.com/nvm-sh/nvm/v0.40.3/install.sh | bash
$ nvm ls-remote
$ nvm install v22.15.0

# Install prerequisites.
$ sudo npm install -g vsce
$ sudo npm install -g http-server
$ sudo npm install -g typescript

# Clone the repo.
$ git clone https://github.boschdevcloud.com/fsil/dse.sdp.git
$ cd dse.sdp

# Setup the SDP.
$ make
$ make build install

# Build and run the example simulation.
$ cd examples/runnable
$ make build
$ make task

# Check the simulation.
$ make report

# Run the simulation using Simer (note: does not run the packaged FMU)
$ make run

# Delete the simulation and generated files
$ make clean
# Delete all downloaded files (in addition to clean).
$ make cleanall
```

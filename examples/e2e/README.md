# Complete E2E Example


## Introduction


### Simulation Project Layout



## Usage

### Using VS Code



### Using Commandline

#### Toolchain Setup

> Note: the Codespace will have the Toolchains already installed and ready to use.

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
# Build the project.
$ cd examples/e2e/project
$ task -y -v

# Run the Simulation.
dse-simer out/sim
```

# Dynamic Simulation Environment - Simulation Development Platform

[![Open in GitHub Codespaces](https://github.com/codespaces/badge.svg)](https://codespaces.new/boschglobal/dse.sdp?quickstart=1)


## Introduction

Simulation Development Platform for the Dynamic Simulation Environment (DSE) Core Platform.


### Project Structure

```text
L- .devcontainer  Devcontainer used by Codespaces.
  L- Dockerfile   Dockerfile used by Codespaces.
L- dsl            DSL parser (usingChevrotain).
L- lsp            VS Code Language Server.
L- licenses       Third Party Licenses.
```


## Usage

> Hint: Codespaces is known to work with Chrome and Edge browsers. Firefox may prevention operation via Firefox's Setting "Enhanced Tracking Protection" (try setting to Standard to resolve the issue).


### Running ModelC Example Simulations

Start a Codespace, then type the following commands in the terminal window.

```bash
# Check your environment.
$ dse-env
DSE_SIMER_IMAGE=ghcr.io/boschglobal/dse-simer:latest
DSE_MODELC_VERSION=2.1.13

# Setup the examples (will download ModelC examples).
$ make examples
$ ls out/examples/modelc/
benchmark/  binary/  extended/  gateway/  gdb/  minimal/  ncodec/  runtime/  simer/  transform/

# Run a simulation using Simer.
$ dse-simer out/examples/minimal
```

> Hint: Find more information about the Simer [command options here](https://boschglobal.github.io/dse.doc/docs/user/simer/#options).



## Build

## Developer Notes

### Dev Containers Extension for VS Code and WSL

> Note: Dev Containers is not supported by VS Codium (workarounds may still exist).

#### Install VS Code Extensions

1. Install the Dev Containers extension (Ctrl-Shift-X, then search "Dev Containers").

2. Install the WSL extension (Ctrl-Shift-X, then search "WSL").

3. From the Remote Explorer, select a WSL Target and then click `Connect in New Window`. A new VS Code editor will open.

4. Press `F1` to bring up the Command Palette and type `Dev Containers reopen`. You may be prompted to install docker into WSL. Even if already installed, proceed to install Docker again (in your WSL)


WSL indicator shows in the bottom left corner of the VS Code window.
The Remote Explorer, added by the WSL extension, will show available WSL Targets.


#### Start VS Code with a WSL Target (and DevContainer)

**Using WSL Bash Terminal:**

```bash
# Open a new VS Code editor connected to this repo.
cd ~/git/workspace/dse.sdp
code.

# In VS Code, press `F1` and type `Dev Containers reopen`.

# Open a bash terminal to access the DevContainer (i.e. local Codespace).
codespace ➜ /workspaces/dse.sdp (main) $ which task
/usr/local/bin/task
codespace ➜ /workspaces/dse.sdp (main) $
```

**Using VS Code:**

1. Click (bottom left) `Open a Remote Window`, then command `Connect to WSL`.

2. From the _Remote Explorer_, select a _WSL Target_ and then click `Connect in Current Window`.<br/>
  A new VS Code editor will open.

3. Select `Open Folder` and naviage to your dse.sdp repo. Previouly opened repos will be listed in the _Remote Explorer_.


After that, the container will build ... and eventually you will have the Codespace avaiable in your Terminal Window.


### Proxy Setup when running _inside_ a DevContainer

https://docs.docker.com/engine/cli/proxy/

~/git/working/dse.sdp$ cat ~/.docker/config.json



## Additional Resources

* [Chevrotain](https://chevrotain.io/docs/) Parser
* [Dev Containers](https://marketplace.visualstudio.com/items?itemName=ms-vscode-remote.remote-containers) for VS Code
* [Simer](https://boschglobal.github.io/dse.doc/docs/user/simer/) (and [source code](https://github.com/boschglobal/dse.modelc/tree/main/extra/tools/simer))



## Contribute

Please refer to the [CONTRIBUTING.md](./CONTRIBUTING.md) file.



## License

Dynamic Simulation Environment FMI Library is open-sourced under the Apache-2.0 license.
See the [LICENSE](LICENSE) and [NOTICE](./NOTICE) files for details.


### Third Party Licenses

[Third Party Licenses](licenses/)

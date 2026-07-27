# vscode_lsp
Development of a VSCode LSP (custom language) for defining simulations.

## Extension Setup
### Extension Packaging
```bash
$ cd dse.sdp/lsp
$ make
```

### Install the Extension
## Install the Extension

#### WSL VS Code

From within the WSL environment, run:

```bash
make install
```
This installs the extension into the VS Code instance connected to WSL.

#### Windows VS Code

To install the extension into the native Windows VS Code:

1. Build the extension:

   ```bash
   cd dse.sdp/lsp
   make
   ```

2. Open **Visual Studio Code (Windows)**.

3. Open the **Extensions** view (`Ctrl + Shift + X`).

4. Click the **⋯ (More Actions)** menu in the Extensions view.

5. Select **Install from VSIX...**.

6. Browse to and select:

   ```text
   dse.sdp/lsp/out/bin/dse.vsix
   ```

7. Reload VS Code if prompted.

## Live AST View

1. Open the DSE supported file.
2. Click the `Open Preview` menu button in the right top corner of vs code.


### Alternatively
- Press <b>`Ctrl + K V`</b> to open Preview to the side.
- Press <b>`Ctrl + Shift + V`</b> to open Preview.

### Important Note
If you are trying the live ast view locally in VS Code (not in Codespace), you may need to run the following command in PowerShell:
```powershell
cd dse.sdp/dsl
npm install -g --force
```
This will ensure that all necessary global dependencies are installed properly.


## DSE Commands


| Command | Description |
| ------- | ----------- |
| **Build** (`DSE: Build`) | Builds `simulation.yaml` and `Taskfile.yml` from dse file |
| **Check** (`DSE: Check`) | Runs graph reporting to visualize and verify simulation |
| **Run** (`DSE: Run`) | Runs the simulation |
| **Clean** (`DSE: Clean`) | Runs `task: clean` |
| **Cleanall** (`DSE: Cleanall`) | Runs `task: cleanall` |
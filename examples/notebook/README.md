# Notebook Based Simulation Example

## Introduction

A Notebook style interface (Jupyter) is used to run an Open Loop simulation.

The example simulation uses a simple CSV file, presented in a Notebook, to operate an Open Loop simulation which represents an Linear Equation.


## Usage

Install the following VS Code extensions:

1. Install the "Jupyter" extension (Ctrl-Shift-X, then search "Jupyter").


### Notebook in VSCode

Start a DevContainer (in WSL):

1. Open the Jupyter Notebook file: `examples/notebook/open_loop_simulation.ipynb`.
2. When prompted, select the Python interpreter to use (e.g. from WSL).
   You can also manually select an interpreter:
   - Press `Ctrl+Shift+P` to open the Command Palette.
   - Type `Python: Select Interpreter` and choose your preferred Python environment in WSL.
3. Click on any cell to start editing.
4. Use the toolbar at the top of the file to run individual cells or the entire notebook.


### Notebook in Browser

Start a Codespace, then type the following commands in the terminal window.

```bash
# Start the notebook server.
$ jupyter notebook --no-browser --ip=0.0.0.0

```

Next, open the Notebook:

1. Open the Notebook URL (from the above command output) in a browser.
2. Click on any cell to start editing.
3. Use the toolbar at the top of the file to run individual cells or the entire notebook.

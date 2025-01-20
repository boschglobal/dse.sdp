# Notebook Based Simulation Example

## Introduction

A Notebook style interface (Jupyter) is used to run an Open Loop simulation.

The example simulation uses a simple CSV file, presented in a Notebook, to operate an Open Loop simulation which represents an Linear Equation.


## Usage

Install the following VS Code extensions:

1. Install the "Jupyter" extension (Ctrl-Shift-X, then search "Jupyter").
2. Install the "Python" extension (Ctrl-Shift-X, then search "Python").

And then install the following PIP packages (via the Terminal):

```bash
$ pip install \
    ipywidgets \
    asammdf[gui] \
    numpy.typing \
    ;
```


### Notebook in VSCode

Start a DevContainer (in WSL):

1. Open the Jupyter Notebook file: `examples/notebook/open_loop_simulation.ipynb`.
2. Click `Run All` to start the Notebook.
3. If (or when) prompted, select a kernel to run the Notebook with:
   1. Select `Python Environments...`
   2. Select the installed (Codespace/WSL) Python environment, for example 'Python 3.12.1'.
4. Using the toolbar at the top of the Notebook, select `Run All` to run the simulation.
   **The results will appear at the bottom of the Notebook**.
5. (Optional )Click on any cell in the Notebook to edit the simulaiton.

> Note: The Python interpreter can also be selected manuall; press `Ctrl+Shift+P` to open the Command Palette, and then type `Python: Select Interpreter` and choose your preferred Python environment in WSL.


### Notebook in Browser

Start a Codespace, then type the following commands in the terminal window.

```bash
# Start the notebook server.
$ cd examples/notebook
$ jupyter notebook --no-browser --ip=0.0.0.0
[I 15:21:09.858 NotebookApp] Writing notebook server cookie secret to /home/codespace/.local/share/jupyter/runtime/notebook_cookie_secret
[I 15:21:10.087 NotebookApp] Serving notebooks from local directory: /workspaces/dse.sdp/examples/notebook
[I 15:21:10.087 NotebookApp] The Jupyter Notebook is running at:
[I 15:21:10.087 NotebookApp] http://codespaces-2855a0:8888/?token=feb7ded293f8890565e514331728d261f3d28bc690a1d39c
[I 15:21:10.087 NotebookApp]  or http://127.0.0.1:8888/?token=feb7ded293f8890565e514331728d261f3d28bc690a1d39c
[I 15:21:10.087 NotebookApp] Use Control-C to stop this server and shut down all kernels (twice to skip confirmation).
[C 15:21:10.089 NotebookApp]

    To access the notebook, open this file in a browser:
        file:///home/codespace/.local/share/jupyter/runtime/nbserver-4999-open.html
    Or copy and paste one of these URLs:
        http://codespaces-2855a0:8888/?token=feb7ded293f8890565e514331728d261f3d28bc690a1d39c
     or http://127.0.0.1:8888/?token=feb7ded293f8890565e514331728d261f3d28bc690a1d39c
```

Next, open the Notebook:

1. Open the Notebook URL (from the above command output) in a browser.
2. Using the toolbar at the top of the Notebook, select `Run All` to run the simulation.
   **The results will appear at the bottom of the Notebook**.
3. (Optional )Click on any cell in the Notebook to edit the simulaiton.

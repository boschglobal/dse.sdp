# vscode_lsp
Development of a VSCode LSP (custom language) for defining simulations.

## Extension Packaging
```bash
$ cd dse.sdp/lsp
$ make
```

## Install the Extension
```bash
$ code --install-extension <path-to-your-extension>.vsix
```

## Extension Activation
Extension gets activated for files with below extensions
```bash
.dse
```

# Live AST View

#### 1. Open the DSE supported file.
#### 2. Click the `Open Preview` menu button in the right top corner of vs code.


### Alternatively:
- Press <b>`Ctrl + K V`</b> to open Preview to the side.
- Press <b>`Ctrl + Shift + V`</b> to open Preview.


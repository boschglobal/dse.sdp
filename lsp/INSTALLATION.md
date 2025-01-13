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
.fsil
.fs
```
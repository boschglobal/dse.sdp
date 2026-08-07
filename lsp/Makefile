# Copyright 2024 Robert Bosch GmbH
#
# SPDX-License-Identifier: Apache-2.0

SCRIPTS_DIR	:= $(CURDIR)/scripts

default: build

.PHONY: build
build:
	mkdir -p out/bin
	mkdir -p bin/graph/yaml
	chmod +x $(SCRIPTS_DIR)/dse-mdf2csv.sh
	ln -sf $(SCRIPTS_DIR)/dse-mdf2csv.sh ~/.local/bin/dse-mdf2csv
	npm install

EXTENSION_NAME := dse.dse-1.0.0
.PHONY: install
install: build
	@if ! command -v vsce >/dev/null 2>&1; then \
		echo "Installing @vscode/vsce..."; \
		npm install -g @vscode/vsce; \
	fi
	vsce package --out ./out/bin/dse.vsix
	$(MAKE) install-extension
	@pip install matplotlib || true
	
.PHONY: install-extension
install-extension:
	@if [ "$(CODESPACES)" = "true" ]; then \
		echo "Installing extension in Codespaces..."; \
		rm -rf /home/vscode/.vscode-remote/extensions/$(EXTENSION_NAME); \
		rm -rf /home/vscode/.vscode-remote/extensions/temp; \
		mkdir -p /home/vscode/.vscode-remote/extensions/temp; \
		unzip -q ./out/bin/dse.vsix "extension/*" -d /home/vscode/.vscode-remote/extensions/temp; \
		mkdir -p /home/vscode/.vscode-remote/extensions/$(EXTENSION_NAME); \
		mv /home/vscode/.vscode-remote/extensions/temp/extension/* /home/vscode/.vscode-remote/extensions/$(EXTENSION_NAME)/; \
		rm -rf /home/vscode/.vscode-remote/extensions/temp; \
	else \
		echo y | code --install-extension ./out/bin/dse.vsix; \
	fi
	@exit 0

.PHONY: test
test:

.PHONY: clean
clean:
	rm -rf out client/out server/out
	rm -rf package-lock.json
	rm -rf client/package-lock.json
	rm -rf server/package-lock.json
	rm -rf bin

.PHONY: cleanall
cleanall: clean
	rm -rf node_modules
	rm -rf client/node_modules
	rm -rf server/node_modules

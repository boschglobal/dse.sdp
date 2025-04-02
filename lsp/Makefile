# Copyright 2024 Robert Bosch GmbH
#
# SPDX-License-Identifier: Apache-2.0

default: build

.PHONY: build
build:
	mkdir -p out/bin
	mkdir -p bin/graph/yaml
	npm install && npm install -g vsce && npm install -g http-server
	vsce package --out ./out/bin/dse.vsix

EXTENSION_NAME := dse.dse-1.0.0
.PHONY: install
install: build
	@if [ "$(CODESPACES)" = "true" ]; then \
		rm -rf /home/codespace/.vscode-remote/extensions/$(EXTENSION_NAME);	\
		rm -rf /home/codespace/.vscode-remote/extensions/temp;	\
		mkdir -p /home/codespace/.vscode-remote/extensions/temp;	\
		unzip -q ./out/bin/dse.vsix "extension/*" -d /home/codespace/.vscode-remote/extensions/temp;	\
		mkdir -p /home/codespace/.vscode-remote/extensions/$(EXTENSION_NAME);	\
		mv /home/codespace/.vscode-remote/extensions/temp/extension/* /home/codespace/.vscode-remote/extensions/$(EXTENSION_NAME)/;	\
		rm -rf /home/codespace/.vscode-remote/extensions/temp;	\
	else \
		echo y | code --install-extension ./out/bin/dse.vsix; \
	fi
	@exit 0

.PHONY: clean
clean:
	rm -rf out
	rm -rf package-lock.json
	rm -rf node_modules

.PHONY: cleanall
cleanall: clean

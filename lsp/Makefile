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
	npm install && npm install -g vsce && npm install -g http-server
	vsce package --out ./out/bin/dse.vsix
	@if [ -f "$$HOME/.bashrc" ]; then \
		grep -qs 'export PATH="$$HOME/.local/bin:$$PATH"' "$$HOME/.bashrc" || \
		echo 'export PATH="$$HOME/.local/bin:$$PATH"' >> "$$HOME/.bashrc"; \
	fi
	@PATH="$$HOME/.local/bin:$$PATH"
	
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
	pip install matplotlib
	@exit 0

.PHONY: test
test:

.PHONY: clean
clean:
	rm -rf out client/out server/out
	rm -rf package-lock.json
	rm -rf node_modules

.PHONY: cleanall
cleanall: clean

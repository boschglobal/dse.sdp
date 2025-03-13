# Copyright 2024 Robert Bosch GmbH
#
# SPDX-License-Identifier: Apache-2.0

default: build

.PHONY: build
build:
	mkdir -p out
	mkdir -p out/bin
	npm install && npm install -g vsce && npm install -g http-server
	vsce package --out ./out

.PHONY: install
install: build
#	code --install-extension out .vsix

.PHONY: clean
clean:
	rm -rf out
	rm -rf package-lock.json
	rm -rf node_modules

.PHONY: cleanall
cleanall: clean

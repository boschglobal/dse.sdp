# Copyright 2024 Robert Bosch GmbH
#
# SPDX-License-Identifier: Apache-2.0


###############
## Docker Images.
TESTSCRIPT_IMAGE ?= ghcr.io/boschglobal/dse-testscript:latest


###############
## Test Parameters.
export HOST_DOCKER_WORKSPACE ?= $(shell pwd -P)
export HOST_ENTRYDIR ?= $(shell pwd -P)
export TESTSCRIPT_E2E_DIR ?= tests/testscript
TESTSCRIPT_E2E_FILES = $(wildcard $(TESTSCRIPT_E2E_DIR)/*.txtar)


default: build


.PHONY: build
build:
	npm install
	npx tsc -b


.PHONY: install
install: build
	export NPM_CONFIG_PREFIX=~/.local; npm install; npm link


.PHONY: test
test: do-test_testscript

do-test_testscript:
# Test debug;
#   Additional logging: add '-v' to Testscript command (e.g. $(TESTSCRIPT_IMAGE) -v \).
#   Retain work folder: add '-work' to Testscript command (e.g. $(TESTSCRIPT_IMAGE) -work \).
	@set -eu; for t in $(TESTSCRIPT_E2E_FILES) ;\
	do \
		echo "Running Test: $$t" ;\
		testscript \
			-e ENTRYDIR=$(HOST_ENTRYDIR) \
			-e REPODIR=$(HOST_DOCKER_WORKSPACE) \
			$$t ;\
	done;


.PHONY: clean
clean:
	rm -f package-lock.json
	rm -rf node_modules


.PHONY: cleanall
cleanall: clean

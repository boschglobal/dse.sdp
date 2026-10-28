# Copyright 2024 Robert Bosch GmbH
#
# SPDX-License-Identifier: Apache-2.0


###############
## Docker Images.
TESTSCRIPT_IMAGE ?= ghcr.io/boschglobal/dse-testscript:latest


###############
## Docker Containers.
DSE_LANG_IMAGE ?= dse-lang


###############
## Test Parameters.
export HOST_DOCKER_WORKSPACE ?= $(shell pwd -P)
export HOST_ENTRYDIR ?= $(shell pwd -P)
export TESTSCRIPT_E2E_DIR ?= tests/testscript
TESTSCRIPT_E2E_FILES = $(wildcard $(TESTSCRIPT_E2E_DIR)/*.txtar)


default: build


.PHONY: build
build: node_modules/.stamp
	npm run compile

node_modules/.stamp: package.json
	npm install --no-audit --no-fund --prefer-offline
	touch $@

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


.PHONY: docker
docker:
	docker build -f Dockerfile -t $(DSE_LANG_IMAGE):test .


.PHONY: clean
clean:
	rm -f package-lock.json
	rm -rf node_modules
	rm -rf out
	docker images -q $(DSE_LANG_IMAGE) | xargs -r docker rmi -f


.PHONY: cleanall
cleanall: clean

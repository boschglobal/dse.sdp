# Copyright 2024 Robert Bosch GmbH
#
# SPDX-License-Identifier: Apache-2.0


################
## DSE Projects.
DSE_MODELC_REPO ?= https://github.com/boschglobal/dse.modelc
DSE_MODELC_VERSION ?= 2.2.12
export DSE_MODELC_PKG_URL ?= $(DSE_MODELC_REPO)/releases/download/v$(DSE_MODELC_VERSION)/ModelC-$(DSE_MODELC_VERSION)-linux-amd64.zip


###############
## Docker Images.
TESTSCRIPT_IMAGE ?= ghcr.io/boschglobal/dse-testscript:latest
SIMER_IMAGE ?= ghcr.io/boschglobal/dse-simer:$(DSE_MODELC_VERSION)


###############
## Build parameters.
SUBDIRS = ast graph dsl lsp doc examples/models


###############
## Test Parameters.
export EXAMPLE_VERSION ?= 0.8.21
export HOST_ENTRYDIR ?= $(shell pwd -P)
export HOST_DOCKER_WORKSPACE ?= $(shell pwd -P)
export TESTSCRIPT_E2E_DIR ?= tests/e2e
TESTSCRIPT_E2E_FILES = $(wildcard $(TESTSCRIPT_E2E_DIR)/*/*.txtar)


default: build


downloads:
	mkdir -p build/downloads
	cd build/downloads; test -s ModelC-$(DSE_MODELC_VERSION)-linux-amd64.zip || ( curl -fSLO $(DSE_MODELC_PKG_URL) && unzip -q ModelC-$(DSE_MODELC_VERSION)-linux-amd64.zip )


.PHONY: examples
examples: downloads
	mkdir -p out/examples
	test -d out/examples/modelc || ( cp -R build/downloads/ModelC-$(DSE_MODELC_VERSION)-linux-amd64/examples out/examples/modelc )
	$(MAKE) -C examples/models build
	cp examples/models/build/_dist/* out/examples


.PHONY: build
build:
	@for d in $(SUBDIRS); do ($(MAKE) -C $$d build ); done
	mkdir -p out/examples
	cp examples/models/build/_dist/* out/examples


.PHONY: test
test:
	@for d in $(SUBDIRS); do ($(MAKE) -C $$d test ); done

.PHONY: test_e2e
test_e2e: do-test_testscript-e2e

.PHONY: install
install:
	@for d in $(SUBDIRS); do ($(MAKE) -C $$d install ); done


.PHONY: clean
clean:
	@for d in $(SUBDIRS); do ($(MAKE) -C $$d clean ); done
	rm -rf out


.PHONY: cleanall
cleanall: clean
	@for d in $(SUBDIRS); do ($(MAKE) -C $$d cleanall ); done
	rm -rf build

.PHONY: docker
docker:
	$(MAKE) -C graph docker
	docker build -f .devcontainer/Dockerfile-builder --tag dse-builder:test . ;\
	docker build -f .devcontainer/Dockerfile --tag dse-devcontainer:test --build-arg DSE_BUILDER_IMAGE=dse-builder:test . ;\

.PHONY: run_graph
run_graph:
	$(MAKE) -C graph graph

.PHONY: generate
generate:
	$(MAKE) -C doc build


do-test_testscript-e2e:
# Test debug;
#   Additional logging: add '-v' to Testscript command (e.g. $(TESTSCRIPT_IMAGE) -v \).
#   Retain work folder: add '-work' to Testscript command (e.g. $(TESTSCRIPT_IMAGE) -work \).	@-docker kill builder 2>/dev/null ; true
#   To skip tests add 'skip' or 'skip message' at start of txtar file.
	@-docker kill builder 2>/dev/null ; true
	@-docker kill report 2>/dev/null ; true
	@-docker kill simer 2>/dev/null ; true
	@set -eu; \
	for t in $(TESTSCRIPT_E2E_FILES); do \
		echo "Running E2E Test: $$t"; \
		export ENTRYWORKDIR=$$(mktemp -d) ;\
		docker run -i --rm \
			--network=host \
			-e ENTRYHOSTDIR=$(HOST_DOCKER_WORKSPACE) \
			-e ENTRYWORKDIR=$${ENTRYWORKDIR} \
			-v /var/run/docker.sock:/var/run/docker.sock \
			-v $(HOST_DOCKER_WORKSPACE):/repo \
			-v $${ENTRYWORKDIR}:/workdir \
			$(TESTSCRIPT_IMAGE) \
				-e ENTRYHOSTDIR=$(HOST_DOCKER_WORKSPACE) \
				-e ENTRYWORKDIR=$${ENTRYWORKDIR} \
				-e REPODIR=/repo \
				-e WORKDIR=/workdir \
				-e DSE_BUILDER_IMAGE=$(DSE_BUILDER_IMAGE) \
				-e DSE_REPORT_IMAGE=$(DSE_REPORT_IMAGE) \
				-e DSE_SIMER_IMAGE=$(DSE_SIMER_IMAGE) \
				-e http_proxy=$(http_proxy) \
				-e https_proxy=$(https_proxy) \
				-e GHE_USER=$(GHE_USER) \
				-e GHE_TOKEN=$(GHE_TOKEN) \
				-e GHE_PAT=$(GHE_PAT) \
				-e AR_USER=$(AR_USER) \
				-e AR_TOKEN=$(AR_TOKEN) \
				-e RELEASE_VERSION=$(EXAMPLE_VERSION) \
				$$t; \
	done


.PHONY: super-linter
super-linter:
	docker run --rm --volume $$(pwd):/tmp/lint \
		--env RUN_LOCAL=true \
		--env DEFAULT_BRANCH=main \
		--env IGNORE_GITIGNORED_FILES=true \
		--env FILTER_REGEX_EXCLUDE="(doc/content/.*|(^|/)vendor/)" \
		--env VALIDATE_CPP=true \
		--env VALIDATE_DOCKERFILE=true \
		--env VALIDATE_MARKDOWN=true \
		--env VALIDATE_YAML=true \
		ghcr.io/super-linter/super-linter:slim-v8

#		--env VALIDATE_GO=true \
#		--env VALIDATE_TYPESCRIPT_ES=true \
#		--env VALIDATE_TYPESCRIPT_PRETTIER=true \

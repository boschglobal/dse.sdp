# Copyright 2025 Robert Bosch GmbH
#
# SPDX-License-Identifier: Apache-2.0

DSE_BUILDER_IMAGE ?= ghcr.io/boschglobal/dse-builder:latest
DSE_REPORT_IMAGE ?= ghcr.io/boschglobal/dse-report:latest
DSE_SIMER_IMAGE ?= ghcr.io/boschglobal/dse-simer:latest

REPO_ROOT := $(shell git rev-parse --show-toplevel)
PROJDIR := /repo/$(shell git rev-parse --show-prefix)

default: build

.PHONY: build
build:
	docker run -it --rm \
		--user $$(id -u):$$(id -g) \
		--group-add $$(stat -c '%g' /var/run/docker.sock) \
		--volume $$(pwd):/workdir \
		--volume $(REPO_ROOT):/repo \
		--workdir $(PROJDIR) \
		-e HOME=/workdir \
		-e PROJDIR=$(PROJDIR) \
		-e WORKDIR=$(PROJDIR) \
		-e ENTRYWORKDIR=$$(pwd) \
		-e WORKDIR=$(PROJDIR) \
		-e AR_USER -e AR_TOKEN -e GHE_USER -e GHE_TOKEN -e GHE_PAT \
		-v /var/run/docker.sock:/var/run/docker.sock \
		$(DSE_BUILDER_IMAGE) openloop.dse

.PHONY: report
report:
	docker run -it --rm \
		--volume $$(pwd)/out/sim:/sim \
		$(DSE_REPORT_IMAGE) report /sim

.PHONY: simer
simer:
	docker run -it --rm \
		--volume $$(pwd)/out/sim:/sim \
		$(DSE_SIMER_IMAGE)

.PHONY: clean
clean:
	@if [ -d out/ ]; then find ./out -mindepth 1 -maxdepth 1 ! -name downloads -exec rm -rf {} +; fi

.PHONY: cleanall
cleanall: clean
	@rm -rf out

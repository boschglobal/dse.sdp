# Copyright 2025 Robert Bosch GmbH
#
# SPDX-License-Identifier: Apache-2.0

export DSE_BUILDER_IMAGE=ghcr.io/boschglobal/dse-builder:latest
export DSE_REPORT_IMAGE=ghcr.io/boschglobal/dse-report:latest
export DSE_SIMER_IMAGE=ghcr.io/boschglobal/dse-simer:latest

export SIM_PATH := $(CURDIR)/out/sim
export ABS_PATH := $(abspath $(CURDIR)/../..)
export LD_LIBRARY_PATH=$(ABS_PATH)/examples/models/build/_deps/dse_modelc_lib-src/examples/runtime/lib
export EXAMPLE_MODELS := $(ABS_PATH)/examples/models
export GATEWAY_PATH := $(EXAMPLE_MODELS)/build/_out/gateway/bin/gateway

export DSE_FILE := gateway.dse
export STACK := model

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
		-e ENTRYWORKDIR=$$(pwd) \
		-e WORKDIR=$(PROJDIR) \
		-e AR_USER -e AR_TOKEN -e GHE_USER -e GHE_TOKEN -e GHE_PAT \
		-v /var/run/docker.sock:/var/run/docker.sock \
		$(DSE_BUILDER_IMAGE) $(DSE_FILE)


.PHONY: report
report:
	docker run -it --rm \
		-v $$(pwd)/out/sim:/sim \
		$(DSE_REPORT_IMAGE) /sim


.PHONY: simer
simer:
	docker run --name simer -d --rm \
		-v $(SIM_PATH):/sim \
		--network=host \
		$(DSE_SIMER_IMAGE) \
			-env simbus:SIMBUS_LOGLEVEL=4 \
			-stack $(STACK) \


.PHONY: gateway
gateway:
	$(GATEWAY_PATH) 0.0005 0.005 $(EXAMPLE_MODELS)/gateway/simulation.yaml $(EXAMPLE_MODELS)/gateway/signalgroup.yaml $(EXAMPLE_MODELS)/gateway/model.yaml


.PHONY: clean
clean:
	rm -rf out

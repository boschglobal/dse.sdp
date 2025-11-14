# Copyright 2025 Robert Bosch GmbH
#
# SPDX-License-Identifier: Apache-2.0

export TASK_X_REMOTE_TASKFILES=1
export DSE_BUILDER_IMAGE=ghcr.io/boschglobal/dse-builder:latest
export DSE_REPORT_IMAGE=ghcr.io/boschglobal/dse-report:latest
export DSE_SIMER_IMAGE=ghcr.io/boschglobal/dse-simer:latest

export SIM_PATH := $(CURDIR)/out/sim
export ABS_PATH := $(abspath $(CURDIR)/../..)
export LD_LIBRARY_PATH=$(ABS_PATH)/examples/models/build/_deps/dse_modelc_lib-src/examples/runtime/lib
export GATEWAY_PATH := $(ABS_PATH)/examples/models/build/_out/gateway/bin/gateway

export DSE_FILE := gateway.dse
export STACK := model


default: build
.PHONY: build
build:
	docker run -it --rm \
		-v $$(pwd):/workdir	\
		$(DSE_BUILDER_IMAGE) /workdir/$(DSE_FILE)
	task -y -v


.PHONY: report
report:
	docker run -it --rm \
		-v $$(pwd)/out/sim:/sim \
		$(DSE_REPORT_IMAGE) /sim


.PHONY: simer
simer:
	docker run --name simer -d --rm\
		-v $(SIM_PATH):/sim \
		--network=host \
		$(DSE_SIMER_IMAGE) \
			-env simbus:SIMBUS_LOGLEVEL=4 \
			-stack $(STACK) \


.PHONY: gateway
gateway:
	$(GATEWAY_PATH) 0.0005 0.005 ./gw.yaml


.PHONY: clean
clean:
	rm -rf out
	rm -rf .task
	rm -f -- *.yml *.json $(filter-out signalgroup.yaml gw.yaml,$(wildcard *.yaml))

# Copyright 2025 Robert Bosch GmbH
#
# SPDX-License-Identifier: Apache-2.0

DSE_BUILDER_IMAGE ?= ghcr.io/boschglobal/dse-builder:latest
DSE_REPORT_IMAGE ?= ghcr.io/boschglobal/dse-report:latest
DSE_SIMER_IMAGE ?= ghcr.io/boschglobal/dse-simer:latest

export TASK_X_REMOTE_TASKFILES=1

default: build

.PHONY: build
build:
	docker run -it --rm \
		--user $$(id -u):$$(id -g) \
		--volume $$(pwd):/workdir \
		$(DSE_BUILDER_IMAGE) openloop.dse
	task -y -v

.PHONY: simer
simer:
	docker run -it --rm \
		--volume $$(pwd)/out/sim:/sim \
		$(DSE_SIMER_IMAGE) \
        	-env SIMBUS_LOGLEVEL=5 \
        	-env linear:SIMBUS_LOGLEVEL=4

.PHONY: clean
clean:
	@if [ -d out/ ]; then find out/ -mindepth 1 -maxdepth 1 -depth -path 'out/downloads' -o -print0 | xargs -0 rm -rf; fi

.PHONY: cleanall
cleanall: clean
	@rm -rf out
	@rm -rf .task
	@rm -f openloop.json
	@rm -f openloop.yaml
	@rm -f simulation.yaml
	@rm -f Taskfile.yml
# Copyright 2026 Robert Bosch GmbH
#
# SPDX-License-Identifier: Apache-2.0

SIMER_FLAGS ?= -env glue:SIMBUS_LOGLEVEL=3 $(SIMBUS_TRACE) --endtime 0.002

DSE_BUILDER_IMAGE ?= ghcr.io/boschglobal/dse-builder:latest
DSE_REPORT_IMAGE ?= ghcr.io/boschglobal/dse-report:latest
DSE_SIMER_IMAGE ?= ghcr.io/boschglobal/dse-simer:latest

export TASK_X_REMOTE_TASKFILES=1

DSE_SCRIPT = simulation.dse
DATA_FILES := $(shell find data -type f)
#MODEL_FILES := $(shell find model -type f)


default: help

.PHONY: help
help:
	@echo "Available targets:"
	@echo "  make pull               - Pull/update latest docker images."
	@echo "  make build              - Build the simulation."
	@echo "  make report             - Report on the simulation (static checks)."
	@echo "  make run                - Run the simulation."
	@echo "  make test               - <not available>"
	@echo "  make package            - Create a ZIP package of the simulation."
	@echo "  make publish            - Upload the ZIP package to artifactory."
	@echo "  make clean              - Clean the simulation folder, remove build artefacts."
	@echo "  make cleanall           - Clean and remove all related artefacts (inc. cached downloads)."


.PHONY: pull
pull:
	docker pull $(DSE_BUILDER_IMAGE)
	docker pull $(DSE_REPORT_IMAGE)
	docker pull $(DSE_SIMER_IMAGE)


.build_dse_script.stamp: $(DSE_SCRIPT)
	docker run -it --rm \
		--user $$(id -u):$$(id -g) \
		--volume $$(pwd):/workdir \
		$(DSE_BUILDER_IMAGE) $(DSE_SCRIPT)
	touch $@

.build_run_task.stamp: $(DSE_SCRIPT) $(MODEL_FILES) $(DATA_FILES)
	task -y -v
	touch $@

.PHONY: build
build: .build_dse_script.stamp .build_run_task.stamp


.PHONY: report
report:
	docker run -it --rm \
		--volume $$(pwd)/out/sim:/sim \
		$(DSE_REPORT_IMAGE) report /sim


.PHONY: run
run:
# TODO fetch an FMU importer.
# TODO start the FMU importer.


.PHONY: clean
clean:
	@if [ -d out/ ]; then find out/ -mindepth 1 -maxdepth 1 -depth -path 'out/downloads' -o -print0 | xargs -0 rm -rf; fi
	@rm -f simulation.json
	@rm -f simulation.yaml
	@rm -f simulation.linux.json
	@rm -f simulation.linux.yaml
	@rm -f Taskfile.yml
	@rm -f .*.stamp

.PHONY: cleanall
cleanall: clean
	@rm -rf out
	@rm -rf .task

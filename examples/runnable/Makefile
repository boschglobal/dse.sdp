# Copyright 2025 Robert Bosch GmbH
#
# SPDX-License-Identifier: Apache-2.0

export TASK_X_REMOTE_TASKFILES=1
export DSE_REPORT_IMAGE=ghcr.io/boschglobal/dse-report:latest

default: build
.PHONY: build
build:
	dse-parse2ast runnable.dse runnable.json
	dse-ast convert -input runnable.json -output runnable.yaml
	dse-ast resolve -input runnable.yaml
	dse-ast generate -input runnable.yaml -output .

.PHONY: task
task:
	task -y -v
	ls -R out/sim
	ls -R out/fmu

.PHONY: report
report:
	docker run -it --rm \
		-v $$(pwd)/out/sim:/sim \
		$(DSE_REPORT_IMAGE) /sim

.PHONY: run
run:
	docker run -it --rm \
		-v $$(pwd)/out/sim:/sim \
		$(DSE_SIMER_IMAGE)
	./post_run.sh

.PHONY: clean
clean:
	rm -rf out/sim
	rm -rf out/fmu

.PHONY: cleanall
cleanall:
	git clean -Xdf

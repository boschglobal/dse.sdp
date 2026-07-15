# Copyright 2025 Robert Bosch GmbH
#
# SPDX-License-Identifier: Apache-2.0
export TASK_X_REMOTE_TASKFILES=1
export DSE_REPORT_IMAGE=ghcr.io/boschglobal/dse-report:latest
export DSE_SIMER_IMAGE=ghcr.io/boschglobal/dse-simer:latest

.PHONY: build
build:
	dse-parse2ast openloop.dse openloop.json
	dse-ast convert -input openloop.json -output openloop.yaml
	dse-ast resolve -input openloop.yaml
	dse-ast generate -input openloop.yaml -output .
	mkdir -p out/sim
	task -y -t out/Taskfile.yml

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
	git clean -Xdf

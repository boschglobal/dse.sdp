# Copyright 2025 Robert Bosch GmbH
#
# SPDX-License-Identifier: Apache-2.0

export TASK_X_REMOTE_TASKFILES=1
export DSE_BUILDER_IMAGE=ghcr.io/boschglobal/dse-builder:latest
export DSE_REPORT_IMAGE=ghcr.io/boschglobal/dse-report:latest
export DSE_SIMER_IMAGE=ghcr.io/boschglobal/dse-simer:latest


default: build
.PHONY: build
build:
	docker run -it --rm \
		-e AR_USER -e AR_TOKEN -e GHE_USER -e GHE_TOKEN -e GHE_PAT \
		-v $$(pwd):/workdir \
		$(DSE_BUILDER_IMAGE) openloop.dse
	task -y -v

.PHONY: report
report:
	docker run -it --rm \
		-v $$(pwd)/out/sim:/sim \
		$(DSE_REPORT_IMAGE) /sim

.PHONY: simer
simer:
	docker run -it --rm \
		-v $$(pwd)/out/sim:/sim \
		$(DSE_SIMER_IMAGE) \
			-transport="loopback" \
		

.PHONY: clean
clean:
	find out/ -mindepth 1 -maxdepth 1 -depth -path 'out/downloads' -o -print0 | xargs -0 rm -rf

.PHONY: cleanall
cleanall:
	git clean -Xdf
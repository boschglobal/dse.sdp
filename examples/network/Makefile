# Copyright 2025 Robert Bosch GmbH
#
# SPDX-License-Identifier: Apache-2.0

export DSE_BUILDER_IMAGE=ghcr.io/boschglobal/dse-builder:latest
export DSE_REPORT_IMAGE=ghcr.io/boschglobal/dse-report:latest
export DSE_SIMER_IMAGE=ghcr.io/boschglobal/dse-simer:latest

export NAME_CAN=network_can
export NAME_PDU=network_pdu

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
		-e AR_USER -e AR_TOKEN -e GHE_USER -e GHE_TOKEN -e GHE_PAT \
		-v /var/run/docker.sock:/var/run/docker.sock \
		$(DSE_BUILDER_IMAGE) network.dse


.PHONY: simer
simer:
	docker run -it --rm \
		-v $$(pwd)/out/sim:/sim \
		$(DSE_SIMER_IMAGE) \
        	-env SIMBUS_LOGLEVEL=2


.PHONY: trace_wildcard
trace_wildcard:
	docker run -it --rm \
		-v $$(pwd)/out/sim:/sim \
		$(DSE_SIMER_IMAGE) -valgrind $(NAME_CAN) \
			-env $(NAME_CAN):SIMBUS_LOGLEVEL=2 \
			-env $(NAME_CAN):NCODEC_TRACE_CAN_1=* \
			-env $(NAME_PDU):NCODEC_TRACE_PDU_47=*


.PHONY: trace_framelist
trace_framelist:
	docker run -it --rm \
		-v $$(pwd)/out/sim:/sim \
		$(DSE_SIMER_IMAGE) -valgrind $(NAME_CAN) \
			-env $(NAME_CAN):SIMBUS_LOGLEVEL=2 \
        	-env $(NAME_CAN):NCODEC_TRACE_CAN_1=0x3ea,0x3eb


.PHONY: clean
clean:
	rm -rf out

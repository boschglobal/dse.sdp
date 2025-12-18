# Copyright 2025 Robert Bosch GmbH
#
# SPDX-License-Identifier: Apache-2.0

export TASK_X_REMOTE_TASKFILES=1
export DSE_BUILDER_IMAGE=ghcr.io/boschglobal/dse-builder:latest
export DSE_REPORT_IMAGE=ghcr.io/boschglobal/dse-report:latest
export DSE_SIMER_IMAGE=ghcr.io/boschglobal/dse-simer:latest

export NAME_CAN=network_can
export NAME_PDU=network_pdu

default: build

.PHONY: build
build:
	docker run -it --rm \
		-v $$(pwd):/workdir \
		$(DSE_BUILDER_IMAGE) network.dse
	task -y -v


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
	rm -rf .task
	rm -f -- *.yml *.json $(filter-out signalgroup_can.yaml signalgroup_pdu.yaml,$(wildcard *.yaml))
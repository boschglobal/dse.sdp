# Copyright 2024 Robert Bosch GmbH
#
# SPDX-License-Identifier: Apache-2.0

SUBDIRS := html2img
PWD := $(shell pwd)
STATIC_DIR := static
OUTPUT_IMG := content/docs/user/tools/dsl/syntax_diagrams.png
export DOCKER_IMAGE := html2img
export DOCKER_BUILDER_CMD := docker run --rm -v $(PWD):/work $(DOCKER_IMAGE) $(STATIC_DIR)/syntax_diagrams.html $(OUTPUT_IMG)


default: build


.PHONY: generate
generate:
	@${DOCKER_BUILDER_CMD}

.PHONY: install
install:

.PHONY: test
test:

.PHONY: build
build:
	@for d in $(SUBDIRS); do ($(MAKE) -C $$d build ); done
	$(MAKE) generate

.PHONY: clean
clean:

.PHONY: cleanall
cleanall:

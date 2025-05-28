# Copyright 2024 Robert Bosch GmbH
#
# SPDX-License-Identifier: Apache-2.0

export DOCKER_IMAGE := html2img
export DOCKER_BUILDER_CMD := docker build -t $(DOCKER_IMAGE) .


default: build


.PHONY: build
build:
	go mod tidy
	@${DOCKER_BUILDER_CMD}

# Copyright 2025 Robert Bosch GmbH
#
# SPDX-License-Identifier: Apache-2.0

export PACKAGE_VERSION ?= 0.0.1
export PACKAGE_ARCH ?= linux-amd64
export PLATFORM_OS := $(word 1,$(subst -, ,$(PACKAGE_ARCH)))
export PLATFORM_ARCH := $(word 2,$(subst -, ,$(PACKAGE_ARCH)))

DSE_MODELC_REPO ?= https://github.com/boschglobal/dse.modelc
DSE_MODELC_VERSION ?= 2.1.32
export DSE_MODELC_LIB_URL ?= $(DSE_MODELC_REPO)/releases/download/v$(DSE_MODELC_VERSION)/ModelC-$(DSE_MODELC_VERSION)-$(PACKAGE_ARCH).zip

GCC_BUILDER_IMAGE ?= ghcr.io/boschglobal/dse-gcc-builder:latest
DOCKER_BUILDER_CMD := docker run -t --rm \
	--user $$(id -u):$$(id -g) \
	--env PACKAGE_ARCH=$(PACKAGE_ARCH) \
	--env PACKAGE_VERSION=$(PACKAGE_VERSION) \
	--env DSE_MODELC_LIB_URL=$(DSE_MODELC_LIB_URL) \
	--volume $$(pwd):/tmp/repo \
	--workdir /tmp/repo \
	$(GCC_BUILDER_IMAGE)

default: build

.PHONY: build
build:
	@${DOCKER_BUILDER_CMD} $(MAKE) do-build

.PHONY: test
test:

.PHONY: clean
clean:
	rm -rf build

.PHONY: cleanall
cleanall: clean

.PHONY: do-build
do-build:
# Build from scratch if no build dir.
	@if [ ! -d "build" ]; then \
		mkdir -p build; \
		cd build; \
		cmake .. ; \
	fi
# Build incremental.
	@cd build; make
	@cd build; make install
	@echo ""
	@echo "Sandbox files: - $$(pwd)/build/_out"
	@echo "--------------"
	@find build/_out/ -type f -name '*' -printf '%06s %p\n'  | numfmt --field=1 --to=iec-i ;
# Package.
	@cd build; make package
	@echo ""
	@echo "Package files: - $$(pwd)/build/_dist"
	@echo "--------------"
	@find build/_dist/ -type f -name '*' -printf '%06s %p\n'  | numfmt --field=1 --to=iec-i ;

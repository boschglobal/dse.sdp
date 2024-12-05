# Copyright 2024 Robert Bosch GmbH
#
# SPDX-License-Identifier: Apache-2.0

export DSE_MODELC_VERSION ?= 2.1.14
export DSE_MODELC_URL ?= https://github.com/boschglobal/dse.modelc/releases/download/v$(DSE_MODELC_VERSION)/ModelC-$(DSE_MODELC_VERSION)-linux-amd64.zip


default: build


downloads:
	mkdir -p build/downloads
	cd build/downloads; test -s ModelC-$(DSE_MODELC_VERSION)-linux-amd64.zip || ( curl -fSLO $(DSE_MODELC_URL) && unzip -q ModelC-$(DSE_MODELC_VERSION)-linux-amd64.zip )


.PHONY: examples
examples: downloads
	mkdir -p out/examples
	test -d out/examples/modelc || ( cp -R build/downloads/ModelC-$(DSE_MODELC_VERSION)-linux-amd64/examples out/examples/modelc )


.PHONY: build
build:


.PHONY: clean
clean:
	rm -rf out


.PHONY: cleanall
cleanall: clean
	rm -rf build

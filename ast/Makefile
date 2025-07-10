# Copyright 2025 Robert Bosch GmbH
#
# SPDX-License-Identifier: Apache-2.0


PACKAGE          := $(shell basename $(shell pwd))
SRC           	 := $(shell find . -type f -name '*.go' -print)
BIN_DIR       	 := $(CURDIR)/bin
CLI_TOOLS    	 ?= $(BIN_DIR)/$(PACKAGE)
GOFLAGS       	 :=
GOPRIVATE     	 ?= github.com/boschglobal


.PHONY: all
all: build

.PHONY: build
build: cmd

cmd: $(CLI_TOOLS)

$(CLI_TOOLS): $(SRC)
	go build $(GOFLAGS) -o $@ ./cmd/$(subst $(BIN_DIR)/,,$@)

.PHONY: install
install: build
	ln -sf $(BIN_DIR)/$(PACKAGE) ~/.local/bin/dse-$(PACKAGE)


.PHONY: test
test:
	go test ./...
	go test ./cmd/... -tags test_e2e

.PHONY: format
format:
	go fmt -x ./...

.PHONY: lint
lint:
	go fmt -x ./...
	docker run -t --rm \
		-v $$(pwd):/app \
		-v ~/.cache/golangci-lint/1.59.1:/root/.cache \
		-w /app \
		golangci/golangci-lint:latest \
		golangci-lint run -v ./...


.PHONY: clean
clean:
	@rm -rf '$(BIN_DIR)'

.PHONY: cleanall
cleanall:

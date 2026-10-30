# Copyright 2025 Robert Bosch GmbH
#
# SPDX-License-Identifier: Apache-2.0


###############
## Build parameters.
PACKAGE         := $(shell basename $(CURDIR))
SRC             := $(shell find . -type f -name '*.go' -print)
BIN_DIR         := $(CURDIR)/bin
CLI_TOOLS       ?= $(BIN_DIR)/$(PACKAGE)
GOFLAGS         ?=
export GOPRIVATE ?= github.com/boschglobal


###############
## Docker.
DOCKER_TOOLS    := report
MEMGRAPH_IMAGE  ?= memgraph/memgraph-mage:3.7.1
MEMGRAPH_NAME   ?= memgraph


default: build
help:
	@echo "Available targets:"
	@echo "  build         Build the graph CLI tool."
	@echo "  install       Install the graph CLI tool and reports to ~/.local."
	@echo "  docker        Build the report images."
	@echo "  docker-mage   Build the report images (memgraph-mage, no cache)."
	@echo "  graph         Start a memgraph container (if not running)."
	@echo "  graph-stop    Stop the memgraph container."
	@echo "  test          Run tests with the race detector (starts a temporary Memgraph with docker)."
	@echo "  test_e2e      Run end-to-end tests (starts Memgraph, see graph)."
	@echo "  format        Format Go source files."
	@echo "  lint          Format and lint Go source files."
	@echo "  clean         Remove build artifacts and report images."
	@echo "  cleanall      Run clean."


.PHONY: build
build: $(CLI_TOOLS)

$(CLI_TOOLS): $(SRC)
	go build $(GOFLAGS) -o $@ ./cmd/$(notdir $@)

.PHONY: install
install: build
	mkdir -p ~/.local/bin ~/.local/share/dse-$(PACKAGE)
	ln -sf $(BIN_DIR)/$(PACKAGE) ~/.local/bin/dse-$(PACKAGE)
	ln -sf $(CURDIR)/cmd/$(PACKAGE)/reports ~/.local/share/dse-$(PACKAGE)/

.PHONY: docker
docker:
	@for d in $(DOCKER_TOOLS); do \
		docker build -f build/package/$$d/Dockerfile \
			--build-arg MEMGRAPH_IMAGE=memgraph \
			--tag dse-$$d:test . || exit 1; \
	done

.PHONY: docker-mage
docker-mage:
	@for d in $(DOCKER_TOOLS); do \
		docker build --no-cache -f build/package/$$d/Dockerfile \
			--build-arg MEMGRAPH_IMAGE=memgraph-mage \
			--tag dse-$$d-mage:test . || exit 1; \
	done


.PHONY: graph
graph:
	@if [ -z "$$(docker container list -q -f name=$(MEMGRAPH_NAME) 2> /dev/null)" ]; then \
		docker run -d --rm --name $(MEMGRAPH_NAME) \
			-p 3000:3000 \
			-p 7444:7444 \
			-p 7687:7687 \
			-v mg_lib:/var/lib/memgraph \
			$(MEMGRAPH_IMAGE); \
	fi

.PHONY: graph-stop
graph-stop:
	docker kill $(MEMGRAPH_NAME)


.PHONY: test
test:
	go test -race -count=1 ./...

.PHONY: test_e2e
test_e2e: graph
	go test -count=1 -p 1 ./cmd/... ./internal/... -tags test_e2e

.PHONY: format
format:
	go fmt ./...

.PHONY: lint
lint: format
	go vet -tags test_e2e ./...
	mkdir -p ~/.cache/golangci-lint
	docker run -t --rm \
		-v $$(pwd):/app \
		-v ~/.cache/golangci-lint:/root/.cache \
		-w /app \
		golangci/golangci-lint:latest \
		golangci-lint run ./...


.PHONY: clean
clean:
	@rm -rf $(BIN_DIR)
	@for d in $(DOCKER_TOOLS); do \
		docker images -q dse-$$d | xargs -r docker rmi -f; \
	done

.PHONY: cleanall
cleanall: clean

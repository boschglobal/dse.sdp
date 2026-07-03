# Copyright 2024 Robert Bosch GmbH
#
# SPDX-License-Identifier: Apache-2.0

# DSE Builder
# ===========
ARG DSE_BUILDER_IMAGE=ghcr.io/boschglobal/dse-builder:latest
FROM ${DSE_BUILDER_IMAGE} AS dse-builder

# Prune build-only DSL dependencies before copying into final image.
FROM dse-builder AS dse-builder-pruned
RUN rm -rf \
    /opt/dse-dsl/node_modules/typescript \
    /opt/dse-dsl/node_modules/tsx \
    /opt/dse-dsl/node_modules/@types \
    /opt/dse-dsl/node_modules/@esbuild \
    /opt/dse-dsl/node_modules/esbuild


# Go Builder
# ==========
FROM golang:bookworm AS go
RUN GOBIN=/usr/local/bin go install github.com/rogpeppe/go-internal/cmd/testscript@latest
RUN GOBIN=/usr/local/bin go install github.com/boschglobal/task/v3/cmd/task@latest


# Construct Final Image
# =====================
FROM mcr.microsoft.com/devcontainers/base:bookworm

ENV DEBIAN_FRONTEND=noninteractive

RUN set -eux; \
    wget https://github.com/mikefarah/yq/releases/latest/download/yq_linux_amd64 -O /usr/bin/yq; \
    chmod +x /usr/bin/yq;

RUN set -eux; \
    apt-get -y update; \
    apt-get -y install --no-install-recommends \
        ca-certificates \
        curl \
        liblua5.4-dev \
        lua5.4 \
        gnupg \
        lsb-release\
    && \
    apt-get clean; \
    rm -rf /var/lib/apt/lists/*;


# Install Go Binaries
# ===================
COPY .devcontainer/.bash_aliases /home/codespace/.bash_aliases
COPY --from=go /usr/local/bin/testscript /usr/local/bin/testscript
COPY --from=go /usr/local/bin/task /usr/local/bin/task


# Install Builder Binaries
# ========================
COPY --from=dse-builder-pruned /usr/local/bin/dse-ast /usr/local/bin/dse-ast
COPY --from=dse-builder-pruned /usr/local/bin/dse-lsp /usr/local/bin/dse-lsp
COPY --from=dse-builder-pruned /opt/dse-dsl/out /opt/dse-dsl/out
COPY --from=dse-builder-pruned /opt/dse-dsl/node_modules /opt/dse-dsl/node_modules
COPY --from=dse-builder-pruned /opt/dse-dsl/package.json /opt/dse-dsl/package.json


# DSL Launch Wrapper
# ==================
RUN echo '#!/bin/sh\nexec node /opt/dse-dsl/out/bin/parse2ast.js "$@"' > /usr/local/bin/dse-parse2ast \
    && chmod +x /usr/local/bin/dse-parse2ast

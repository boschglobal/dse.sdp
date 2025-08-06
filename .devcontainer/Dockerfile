# Copyright 2024 Robert Bosch GmbH
#
# SPDX-License-Identifier: Apache-2.0

# DSE Builder
# ===========
ARG DSE_BUILDER_IMAGE=ghcr.io/boschglobal/dse-builder:latest
FROM $DSE_BUILDER_IMAGE AS dse-builder


# Go Builder
# ==========
FROM golang:bullseye AS go
RUN GOBIN=/usr/local/bin go install github.com/rogpeppe/go-internal/cmd/testscript@latest
RUN GOBIN=/usr/local/bin go install github.com/boschglobal/task/v3/cmd/task@latest


# Construct Final Image
# =====================
FROM mcr.microsoft.com/devcontainers/universal:2

ENV DEBIAN_FRONTEND=noninteractive

RUN set -eux; \
    wget https://github.com/mikefarah/yq/releases/latest/download/yq_linux_amd64 -O /usr/bin/yq; \
    chmod +x /usr/bin/yq;

RUN set -eux; \
    apt-get -y -o Dpkg::Options::="--force-confdef" -o Dpkg::Options::="--force-confold" upgrade; \
    apt-get -y install --no-install-recommends \
        jupyter-notebook \
    && \
    apt-get clean; \
    rm -rf /var/lib/apt/lists/*;

RUN pip install asammdf[gui]


# Install VS Code Server Environment
# ==================================
RUN npm install -g vsce
RUN npm install -g typescript
RUN npm install -g http-server


# Install Go Binaries
# ===================
COPY .devcontainer/.bash_aliases /home/codespace/.bash_aliases
COPY --from=go /usr/local/bin/testscript /usr/local/bin/testscript
COPY --from=go /usr/local/bin/task /usr/local/bin/task


# Install Builder Binaries
# ========================
COPY --from=dse-builder /usr/local/bin/dse-ast /usr/local/bin/dse-ast
COPY --from=dse-builder /usr/local/bin/dse-lsp /usr/local/bin/dse-lsp
COPY --from=dse-builder /opt/dse-dsl/out /opt/dse-dsl/out
COPY --from=dse-builder /opt/dse-dsl/node_modules /opt/dse-dsl/node_modules
COPY --from=dse-builder /opt/dse-dsl/package.json /opt/dse-dsl/package.json


# DSL Launch Wrapper
# ==================
RUN echo '#!/bin/sh\nexec node /opt/dse-dsl/out/bin/parse2ast.js "$@"' > /usr/local/bin/dse-parse2ast \
    && chmod +x /usr/local/bin/dse-parse2ast

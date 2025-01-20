# Copyright 2024 Robert Bosch GmbH
#
# SPDX-License-Identifier: Apache-2.0


# Go Builder
# ==========
FROM golang:bullseye as go
RUN GOBIN=/usr/local/bin go install github.com/rogpeppe/go-internal/cmd/testscript@latest
RUN GOBIN=/usr/local/bin go install github.com/go-task/task/v3/cmd/task@latest


# Construct Final Image
# =====================
FROM mcr.microsoft.com/devcontainers/universal:2
RUN set -eux; \
    wget https://github.com/mikefarah/yq/releases/latest/download/yq_linux_amd64 -O /usr/bin/yq; \
    chmod +x /usr/bin/yq;
RUN set -eux; \
	apt-get -y update; \
	apt-get -y upgrade; \
    apt-get -y install --no-install-recommends \
        jupyter-notebook \
    && \
    apt-get clean; \
    rm -rf /var/lib/apt/lists/*;
COPY .devcontainer/.bash_aliases /home/codespace/.bash_aliases
COPY --from=go /usr/local/bin/testscript /usr/local/bin/testscript
COPY --from=go /usr/local/bin/task /usr/local/bin/task

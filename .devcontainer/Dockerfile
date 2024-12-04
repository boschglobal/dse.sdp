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
COPY .devcontainer/.bash_aliases /home/codespace/.bash_aliases
COPY --from=go /usr/local/bin/testscript /usr/local/bin/testscript
COPY --from=go /usr/local/bin/task /usr/local/bin/task

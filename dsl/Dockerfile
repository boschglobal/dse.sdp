# Copyright 2026 Robert Bosch GmbH
#
# SPDX-License-Identifier: Apache-2.0


# DSE Lang
#
# Usage (from this directory):
#	cat path/to/input.dse | docker run --rm -i dse-lang:test validate -
#	docker run --rm -i dse-lang:test validate - < path/to/input.dse


# Launcher Builder
# ===============
FROM golang:1.23-alpine AS launcher
WORKDIR /src
COPY launcher.go .
# Compile a statically linked binary compatible with distroless (no libc dependencies)
RUN --mount=type=cache,target=/root/.cache/go-build \
    --mount=type=cache,target=/go/pkg/mod \
    CGO_ENABLED=0 GOOS=linux go build -o launcher launcher.go


# Validator Builder
# =================
FROM node:22 AS validator
WORKDIR /app
COPY package*.json ./
RUN --mount=type=cache,target=/root/.npm \
    npm ci --omit=dev --no-audit --no-fund --prefer-offline
COPY validate.ts ./validate.ts
COPY lib ./lib
RUN npm run build:validator


# DSE Lang Container Image
# ========================
FROM gcr.io/distroless/nodejs22-debian13:nonroot
WORKDIR /app
COPY --from=launcher /src/launcher /app/launcher
COPY --from=validator /app/out/validate.cjs /app/dse-lang-compile.cjs
COPY dse-script.ebnf /usr/share/dse-lang/dse-script.ebnf

ENTRYPOINT ["/app/launcher"]

# Copyright 2026 Robert Bosch GmbH
#
# SPDX-License-Identifier: Apache-2.0

# HTML2IMG Builder
# ================
FROM golang:bookworm AS html2img
WORKDIR /go/src/html2img

COPY go.mod go.sum ./
RUN go mod download

COPY main.go ./
RUN CGO_ENABLED=0 go build -o bin/html2img main.go


# HTML2IMG Container Image
# ========================
FROM debian:bookworm-slim
LABEL maintainer="timothy.rule@de.bosch.com"

ENV DEBIAN_FRONTEND=noninteractive

RUN set -eux; \
    apt-get -y update; \
    apt-get -y install --no-install-recommends \
        ca-certificates \
        wget \
    && \
    wget -O /tmp/google-chrome-stable_current_amd64.deb \
        https://dl.google.com/linux/direct/google-chrome-stable_current_amd64.deb; \
    apt-get -y install --no-install-recommends \
        /tmp/google-chrome-stable_current_amd64.deb \
    && \
    rm -f /tmp/google-chrome-stable_current_amd64.deb; \
    apt-get clean; \
    rm -rf /var/lib/apt/lists/*;

COPY --from=html2img /go/src/html2img/bin/html2img /usr/local/bin/html2img

WORKDIR /work
ENTRYPOINT ["html2img"]

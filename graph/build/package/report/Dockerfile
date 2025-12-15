# Copyright 2025 Robert Bosch GmbH
#
# SPDX-License-Identifier: Apache-2.0

ARG MEMGRAPH_IMAGE=memgraph
ARG MEMGRAPH_VERSION=3.7.1


# Graph Builder
# =============
FROM golang:bookworm AS graph
WORKDIR /go/src/graph
COPY . .
RUN go build -o bin/graph ./cmd/graph


# Report Container Image
# ======================
FROM memgraph/${MEMGRAPH_IMAGE}:${MEMGRAPH_VERSION}
LABEL maintainer="timothy.rule@de.bosch.com"

USER root
RUN set -eux; \
    apt-get update; \
    apt-get install -y --no-install-recommends \
        netcat-openbsd \
    && \
    apt-get clean; \
    rm -rf /var/lib/apt/lists/*;

COPY --from=graph /go/src/graph/bin/graph /usr/local/bin/graph
COPY cmd/graph/reports /home/memgraph/.local/share/dse-graph/reports
COPY build/package/report/entrypoint.sh /entrypoint.sh

# Setup working directory
WORKDIR /sim
RUN set -eux; \
    chown -R memgraph:memgraph /sim; \
    chmod +x /entrypoint.sh

# Switch to non-root user
USER memgraph
ENV HOME=/home/memgraph
ENTRYPOINT ["/entrypoint.sh"]
EXPOSE 7687

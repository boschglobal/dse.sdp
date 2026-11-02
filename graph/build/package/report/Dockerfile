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
FROM memgraph/${MEMGRAPH_IMAGE}:${MEMGRAPH_VERSION} AS base
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


# Slim
# ====
# Remove unused analytics packages and modules.
FROM base AS slim-trim
USER root
RUN set -eux; \
    cd /usr/local/lib/python3.12/dist-packages; \
    rm -rf \
        numpy numpy.libs numpy-*.dist-info \
        scipy scipy.libs scipy-*.dist-info \
        gensim gensim-*.dist-info \
        networkx networkx-*.dist-info; \
    cd /usr/lib/memgraph/query_modules; \
    grep -rlE '(numpy|scipy|gensim|networkx)' --include='*.py' . | xargs -r rm -rf; \
    rm -rf node2vec_online_utils; \
    find /usr -name __pycache__ -type d -prune -exec rm -rf {} +

# Flatten so that the deleted files are not retained in lower layers.
FROM scratch AS slim
COPY --from=slim-trim / /
ENV PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin
ENV HOME=/home/memgraph
WORKDIR /sim
VOLUME ["/var/log/memgraph", "/var/lib/memgraph", "/etc/memgraph"]
USER memgraph
ENTRYPOINT ["/entrypoint.sh"]
EXPOSE 7687


# Full
# ====
FROM base AS full

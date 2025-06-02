# Copyright 2025 Robert Bosch GmbH
#
# SPDX-License-Identifier: Apache-2.0


# Graph Builder
# ==================
FROM golang:bookworm AS graph
WORKDIR /go/src/graph
COPY . .
RUN go build -o bin/graph ./cmd/graph


# Graph Container Image
# ================
FROM ubuntu:22.04

ENV DEBIAN_FRONTEND=noninteractive TZ=Etc/UTC

# Install dependencies
RUN apt-get update && apt-get install -y \
    adduser \
    curl \
    libatomic1 \
    libatspi2.0-0 \
    libcurl4 \
    libgtk-3-0 \
    libnotify4 \
    libsecret-1-0 \
    libseccomp2 \
    libssl3 \
    libxmlsec1-dev \
    libxss1 \
    libxtst6 \
    netcat \
    python3 \
    python3-pip \
    tzdata \
    unzip \
    xmlsec1 && \
    rm -rf /var/lib/apt/lists/* /tmp/* /var/tmp/*

# Install Python packages
RUN pip3 install \
    gensim==4.3.3 \
    networkx==3.4.2 \
    numpy==1.26.4 \
    scipy==1.13.0 \
    xmlsec==1.3.14


# Install Memgraph
# ================
RUN curl -L https://download.memgraph.com/memgraph/v3.2.0/ubuntu-22.04/memgraph_3.2.0-1_amd64.deb -o /tmp/memgraph.deb && \
    dpkg -i /tmp/memgraph.deb && rm /tmp/memgraph.deb


# Install graph binary
# ====================
COPY --from=graph /go/src/graph/bin/graph /usr/local/bin/graph


# Install reports
# ===============
COPY cmd/graph/reports /home/memgraph/.local/share/dse-graph/reports
RUN chown -R memgraph:memgraph /home/memgraph/.local


# Set up Memgraph directories and permissions
# ===========================================
RUN mkdir -p /usr/lib/memgraph /var/lib/memgraph /var/log/memgraph /etc/memgraph /opt/memgraph-lab && \
    chown -R memgraph:memgraph /usr/lib/memgraph /var/lib/memgraph /var/log/memgraph /etc/memgraph /opt/memgraph-lab


# Expose Port
# ===========
EXPOSE 7687


# Setup the entry conditions
# ==========================
# Run docker with --volume parameter to map the simulation folder into /sim.
WORKDIR /sim
COPY build/package/report/entrypoint.sh /entrypoint.sh
RUN chmod +x /entrypoint.sh

# Switch to the non-root user
USER memgraph

# Set HOME for the memgraph user
ENV HOME=/home/memgraph

ENTRYPOINT ["/entrypoint.sh"]

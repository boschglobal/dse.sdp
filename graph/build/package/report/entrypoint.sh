#!/bin/bash

# Copyright 2025 Robert Bosch GmbH
#
# SPDX-License-Identifier: Apache-2.0

set -e

: "${GRAPH_EXE:=/usr/local/bin/graph}"

function print_usage () {
    echo "Report Command"
    echo ""
    echo "dse-report <sim-dir> "
    echo ""
    echo "Examples:"
    echo 'dse-report examples/graph/stack/sim_good'
    echo 'dse-report -name="Model UID Check;ModelInstance Name Check" examples/graph/stack/sim_good'
    echo 'dse-report -tag=stack examples/graph/stack/sim_good'
    echo 'dse-report -list'
}

# Print usage if no arguments are provided.
if [ $# -eq 0 ]; then
    print_usage
    exit 0
fi

# Start Memgraph ...
/usr/lib/memgraph/memgraph \
  --log-level=ERROR \
  --query-modules-directory=/usr/lib/memgraph/query_modules \
  </dev/null > /dev/null 2> >(grep -i 'error' >&2) & disown
# Wait for Memgraph to be ready ...
for i in {1..60}; do
  if (echo > /dev/tcp/localhost/7687) >/dev/null 2>&1; then
    break
  fi
  echo "Waiting for Memgraph to be ready..."
  sleep 0.5
done
# Verify Memgraph actually started.
if ! nc -z localhost 7687; then
  echo "ERROR: Memgraph failed to start!"
  exit 1
fi

# Continue argument processing.
SKIP=false
for arg in "$@"; do
  case "$arg" in
    -list|-list-all|-list-tags|-h|help)
      SKIP=true
      break
      ;;
  esac
done
if [ "$SKIP" = false ]; then
  $GRAPH_EXE ping > /dev/null 2>&1
  $GRAPH_EXE drop --all > /dev/null 2>&1
fi

# Run the report command.
exec $GRAPH_EXE "$@"

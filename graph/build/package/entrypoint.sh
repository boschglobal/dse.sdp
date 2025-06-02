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

# Print usage if no arguments are provided
if [ $# -eq 0 ]; then
    print_usage
    exit 0
fi

# Start Memgraph
/usr/lib/memgraph/memgraph --log-level=ERROR </dev/null > /dev/null 2> >(grep -i 'error' >&2) & disown

# Wait for Memgraph to be ready
for i in {1..15}; do
  if nc -z localhost 7687; then
    break
  fi
  sleep 1
done

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

# Run report command
exec $GRAPH_EXE report "$@"

#!/usr/bin/env bash
# shellcheck shell=bash

# Usage:
#   source ./setup.sh
#   source <(curl -fsSL https://raw.githubusercontent.com/boschglobal/dse.sdp/main/setup.sh)
# What it does:
# - exports DSE image variables
# - exports TASK_X_REMOTE_TASKFILES=1
# - defines convenience aliases:
#     dse-builder
#     dse-simer
#     dse-trace

if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
  echo "Please source this script instead of executing it:"
  echo "  source ./setup.sh"
  exit 1
fi

export DSE_BUILDER_IMAGE="${DSE_BUILDER_IMAGE:-ghcr.io/boschglobal/dse-builder:latest}"
export DSE_SIMER_IMAGE="${DSE_SIMER_IMAGE:-ghcr.io/boschglobal/dse-simer:latest}"
export TASK_X_REMOTE_TASKFILES="${TASK_X_REMOTE_TASKFILES:-1}"

alias dse-builder='docker run -it --rm --user $(id -u):$(id -g) -v "$(pwd)":/workdir "${DSE_BUILDER_IMAGE}"'

_dse_simer() {
  local sim_dir="out/sim"

  if [[ $# -gt 0 && ! -d "$1" ]]; then
    sim_dir="$1"
    shift
  fi

  docker run -it --rm \
    -v "$(pwd)/${sim_dir}:/sim" \
    "${DSE_SIMER_IMAGE}" \
    "$@"
}

_dse_trace() {
  local sim_dir="out/sim"

  if [[ $# -gt 0 && ! -d "$1" ]]; then
    sim_dir="$1"
    shift
  fi

  docker run -it --rm \
    --entrypoint "" \
    --user "$(id -u):$(id -g)" \
    -v "$(pwd)/${sim_dir}:/sim" \
    "${DSE_SIMER_IMAGE}" \
    /usr/local/bin/trace "$@"
}

alias dse-simer='_dse_simer'
alias dse-trace='_dse_trace'

echo "DSE SDP environment configured."
echo "Images:"
echo "  DSE_BUILDER_IMAGE=${DSE_BUILDER_IMAGE}"
echo "  DSE_SIMER_IMAGE=${DSE_SIMER_IMAGE}"
echo
echo "Available commands:"
echo "  dse-builder simulation.dse"
echo "  dse-simer [sim-path]"
echo "  dse-trace [sim-path] convert --csv data/simbus.bin"

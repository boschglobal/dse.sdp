#!/usr/bin/env bash

INPUT_DSE="$1"
export TASK_X_REMOTE_TASKFILES=1
export DSE_BUILDER_IMAGE=ghcr.io/boschglobal/dse-builder:latest
export DSE_REPORT_IMAGE=ghcr.io/boschglobal/dse-report:latest
export DSE_SIMER_IMAGE=ghcr.io/boschglobal/dse-simer:latest

run_builder() {
  echo "[INFO] Running builder..."
  docker run --rm \
    --network=host \
    -e AR_USER -e AR_TOKEN -e GHE_USER -e GHE_TOKEN -e GHE_PAT \
    -e http_proxy -e https_proxy -e no_proxy \
    -v "$(pwd)":/workdir \
    --user "$(id -u):$(id -g)" \
    "$DSE_BUILDER_IMAGE" "$INPUT_DSE"
  task -y -v
}

if echo " $* " | grep -q " BUILD_ONLY "; then
  run_builder
else  
  run_builder

  echo "[INFO] Running report..."
  docker run --rm \
    -v "$(pwd)/out/sim":/sim \
    "$DSE_REPORT_IMAGE" /sim

  echo "[INFO] Running simer..."
  docker run --rm \
    -v "$(pwd)/out/sim":/sim \
    -p 2159:2159 -p 6379:6379 \
    "$DSE_SIMER_IMAGE" \
    -redis="" \
    -simbus="" \
    -transport="loopback" \
    -uri="loopback" \
    -endtime 0.100 \
    -logger 4 \
    -env SIMBUS_LOGLEVEL=2
fi
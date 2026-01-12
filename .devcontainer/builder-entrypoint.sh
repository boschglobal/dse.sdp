#!/bin/bash
# Copyright 2025 Robert Bosch GmbH
#
# SPDX-License-Identifier: Apache-2.0

set -euo pipefail

DSE_FILE="$1"
SIM_NAME="${DSE_FILE%.*}"

echo "Building simulation: $SIM_NAME"

# Builder part.
dse-parse2ast "$DSE_FILE" "${SIM_NAME}.json"
dse-ast convert -input "${SIM_NAME}.json" -output "${SIM_NAME}.yaml"
dse-ast resolve -input "${SIM_NAME}.yaml"
dse-ast generate -input "${SIM_NAME}.yaml" -output .

# Ensure sim folder is created (required by Task).
mkdir -p out/sim

# Run Task.
TASK_X_REMOTE_TASKFILES=1
# TODO: task -y -v
# TODO: DinD volume mapping etc.
# TODO: Control by a parameter? Or always run task?
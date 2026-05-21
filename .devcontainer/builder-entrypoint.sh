#!/bin/bash
# Copyright 2025 Robert Bosch GmbH
#
# SPDX-License-Identifier: Apache-2.0

set -euo pipefail

DSE_FILE="$1"
SIM_NAME="${DSE_FILE%.*}"
OVERWRITE="${OVERWRITE:-0}"
OVERWRITE_FLAG=""

case "$(printf '%s' "$OVERWRITE" | tr '[:upper:]' '[:lower:]')" in
    1|true)
        OVERWRITE_FLAG="--overwrite"
        ;;
    0|false)
        OVERWRITE_FLAG=""
        ;;
    *)
        echo "Invalid OVERWRITE value: $OVERWRITE"
        echo "Accepted values: 0, 1, true, false"
        exit 1
        ;;
esac

echo "Building simulation: $SIM_NAME"
echo "Overwrite mode: $OVERWRITE"

# Builder part.
dse-parse2ast "$DSE_FILE" "${SIM_NAME}.json"
dse-ast convert -input "${SIM_NAME}.json" -output "${SIM_NAME}.yaml"
dse-ast resolve -input "${SIM_NAME}.yaml"
dse-ast generate -input "${SIM_NAME}.yaml" -output . --script "$DSE_FILE" $OVERWRITE_FLAG

# Ensure sim folder is created (required by Task).
mkdir -p out/sim
chmod 777 out/sim

# Run Task.
TASK_X_REMOTE_TASKFILES=1
# TODO: task -y -t out/Taskfile.yml
# TODO: DinD volume mapping etc.
# TODO: Control by a parameter? Or always run task?
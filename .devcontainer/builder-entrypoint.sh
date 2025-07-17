#!/bin/bash
# Copyright 2025 Robert Bosch GmbH
#
# SPDX-License-Identifier: Apache-2.0

set -euo pipefail

DSE_FILE="$1"
SIM_NAME="${DSE_FILE%.*}"

echo "Building simulation: $SIM_NAME"

dse-parse2ast "$DSE_FILE" "${SIM_NAME}.json"
dse-ast convert -input "${SIM_NAME}.json" -output "${SIM_NAME}.yaml"
dse-ast resolve -input "${SIM_NAME}.yaml"
dse-ast generate -input "${SIM_NAME}.yaml" -output .

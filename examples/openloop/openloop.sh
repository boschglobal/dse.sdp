#!/bin/bash

# Copyright 2024 Robert Bosch GmbH
#
# SPDX-License-Identifier: Apache-2.0

export TASK_X_REMOTE_TASKFILES=1
export SIMER_IMAGE=ghcr.io/boschglobal/dse-simer:latest
export SIM_DIR=out/sim
export MDF_FILE=measurement.mf4

simer() { ( cd "$1" && shift && docker run -it --rm -v $(pwd):/sim $SIMER_IMAGE "$@"; ) }

task -y build
simer $SIM_DIR \
    -env linear:MEASUREMENT_FILE=/sim/$MDF_FILE \
    -stepsize 0.0005 -endtime 0.005

echo ""
echo "Simulation complete."
echo "Measurement file : $SIM_DIR/$MDF_FILE"

python3 - <<'__PY_MF42CSV'
import os
from asammdf import MDF

in_file = os.path.expandvars("$SIM_DIR/$MDF_FILE")
out_file = os.path.expandvars("$SIM_DIR/measurement.csv")
mdf = MDF(in_file)
mdf.export(fmt="csv", filename=out_file)
__PY_MF42CSV

echo "Measurement file : $SIM_DIR/measurement.ChannelGroup_0_linear.csv (converted)"
echo ""

exit 0

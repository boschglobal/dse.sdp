#!/bin/bash

# Copyright 2024 Robert Bosch GmbH
#
# SPDX-License-Identifier: Apache-2.0

# Input Parameters
# ================
#   SRCDIR      - Source directory (of this script and assocaited files).
#                 (optional)
#   WORKDIR     - Working directory, where the simulaiton artifacts are
#                 downloaded and the simulation is created.
#                 (optional)
#   INPUTCSV    -

: "${SRCDIR:=$(pwd)}"
: "${WORKDIR:=$(mktemp -d)}"


echo "Open Loop Simulation"
echo "===================="
echo ""
echo "Environment:"
echo "------------"
echo "SRCDIR=${SRCDIR}"
echo "WORKDIR=${WORKDIR}"
echo "INPUTCSV=${INPUTCSV}"


# Taskfile


# Copy CSV to simulation folder.


# Simer





echo "done"

exit 0

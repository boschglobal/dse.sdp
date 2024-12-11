#!/bin/bash

# Copyright 2024 Robert Bosch GmbH
#
# SPDX-License-Identifier: Apache-2.0


[ $# -eq 0 ] && echo "incorrect number of arguments" && exit 1
[ ! -f "$1" ] && echo "argument is not regular file" && exit 1

AST_FILE=$1

SIM_COUNT=`jq  '[.. | objects | select(.object.payload.simulation_arch)] | length' $AST_FILE`
CH_COUNT=`jq  '[.. | objects | select(.object.payload.channel_name)] | length' $AST_FILE`
NET_COUNT=`jq  '[.. | objects | select(.object.payload.network_name)] | length' $AST_FILE`
USES_COUNT=`jq  '[.. | objects | select(.object.payload.use_item)] | length' $AST_FILE`
MODEL_COUNT=`jq  '[.. | objects | select(.object.payload.model_name)] | length' $AST_FILE`
STACK_COUNT=`jq  '[.. | objects | select(."fmu-stacked".payload.stack_name)] | length' $AST_FILE`

printf "Statistics for file : %s\n" $AST_FILE
printf "sims = %s\n" $SIM_COUNT
printf "channels = %s\n" $CH_COUNT
printf "networks = %s\n" $NET_COUNT
printf "uses = %s\n" $USES_COUNT
printf "models = %s\n" $MODEL_COUNT
printf "stacks = %s\n" $STACK_COUNT

exit 0

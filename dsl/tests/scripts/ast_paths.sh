#!/bin/bash

# Copyright 2024 Robert Bosch GmbH
#
# SPDX-License-Identifier: Apache-2.0


[ $# -eq 0 ] && echo "incorrect number of arguments" && exit 1
[ ! -f "$1" ] && echo "argument is not regular file" && exit 1

AST_FILE=$1

jq --stream -r 'select(.[1]|scalars!=null) | "\(.[0]|join(".")): \(.[1]|tojson) :"' $AST_FILE

exit 0

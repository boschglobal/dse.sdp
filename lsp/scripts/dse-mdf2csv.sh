#!/bin/bash

if [ $# -ne 1 ]; then
	echo "Error: Missing required argument."
    echo "Usage: $(basename "$0") <mf4_file>"
    exit 1
fi

MF4_FILE=$1
BASE_NAME=$(basename "$MF4_FILE" .mf4)
CSV_FILE="${BASE_NAME}.csv"
FULL_PATH=$(realpath "$CSV_FILE")

python3 - <<EOF
import os
import matplotlib.pyplot as plt
from asammdf import MDF

mdf = MDF("$MF4_FILE")
df = mdf.to_dataframe()
df.to_csv("$CSV_FILE")
print("✅ CSV file created: $FULL_PATH")
EOF

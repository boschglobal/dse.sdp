#!/bin/bash

if [ $# -ne 1 ]; then
	echo "Error: Missing required argument."
    echo "Usage: $(basename "$0") <mf4_file>"
    exit 1
fi

MF4_FILE=$1

if [ ! -f "$MF4_FILE" ]; then
    echo "Error: File '$MF4_FILE' does not exist."
    exit 1
fi

CSV_PATH="$(dirname "$MF4_FILE")"/"$(basename "$MF4_FILE" .mf4).csv"

python3 - <<EOF
import os
import matplotlib.pyplot as plt
from asammdf import MDF

mdf = MDF("$MF4_FILE")
df = mdf.to_dataframe()
df.to_csv("$CSV_PATH")
print("✅ CSV file created: $CSV_PATH")
EOF

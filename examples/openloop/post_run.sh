export SIM_DIR=out/sim
export MDF_FILE=measurement.mf4
echo ""
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
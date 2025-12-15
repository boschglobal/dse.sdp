#!/usr/bin/env bash

# To use local images: export DSE_REPORT_IMAGE=dse-report:test
: "${DSE_REPORT_IMAGE:=ghcr.io/boschglobal/dse-report:latest}"

echo "Graph Script"
echo "============"
echo "DSE_REPORT_IMAGE=${DSE_REPORT_IMAGE}"
echo ""

echo "[INFO] Running report..."

docker run --name report -i --rm \
    --network=host \
    -v $ENTRYHOSTDIR:/repo \
    -v $WORKDIR:/workdir \
    $DSE_REPORT_IMAGE $@

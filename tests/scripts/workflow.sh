#!/usr/bin/env bash

INPUT_DSE="$1"

# To use local images: export DSE_BUILDER_IMAGE=dse-builder:test
: "${DSE_BUILDER_IMAGE:=ghcr.io/boschglobal/dse-builder:latest}"
: "${DSE_REPORT_IMAGE:=ghcr.io/boschglobal/dse-report:latest}"
: "${DSE_SIMER_IMAGE:=ghcr.io/boschglobal/dse-simer:latest}"


echo "Workflow Script"
echo "==============="
echo "DSE_BUILDER_IMAGE=${DSE_BUILDER_IMAGE}"
echo "DSE_REPORT_IMAGE=${DSE_REPORT_IMAGE}"
echo "DSE_SIMER_IMAGE=${DSE_SIMER_IMAGE}"
echo ""


export TASK_X_REMOTE_TASKFILES=1
export DSE_BUILDER_IMAGE
export DSE_REPORT_IMAGE
export DSE_SIMER_IMAGE


run_builder() {
    echo "[INFO] Running builder..."
    # Expand envars.
    envsubst < "$INPUT_DSE" > "${INPUT_DSE}.tmp"
    mv "${INPUT_DSE}.tmp" "$INPUT_DSE"
    # Run the Builder.
    docker run --name builder -i --rm \
        --network=host \
        --user "$(id -u):$(id -g)" \
        -v $ENTRYWORKDIR:/workdir \
        -v $ENTRYHOSTDIR:/repo \
        -e AR_USER -e AR_TOKEN -e GHE_USER -e GHE_TOKEN -e GHE_PAT \
        -e http_proxy -e https_proxy -e no_proxy \
        $DSE_BUILDER_IMAGE "$INPUT_DSE"
    # Run Task to finalize the simulation.
    echo "[INFO] Running task..."
    task -y -v
}

run_report() {
    echo "[INFO] Running report..."
    # Graph container should be running, either locally or as an attached
    # container in a CI workflow.
    docker run --name report -i --rm \
        --network=host \
        -v $ENTRYWORKDIR/out/sim:/sim \
        $DSE_REPORT_IMAGE /sim
}

run_simer() {
    echo "[INFO] Running simer..."
    docker run --name simer -i --rm \
        --network=host \
        --user "$(id -u):$(id -g)" \
        -v $ENTRYWORKDIR/out/sim:/sim \
        -p 6379:6379 \
        $DSE_SIMER_IMAGE \
            -env SIMBUS_LOGLEVEL=2
}

if echo " $* " | grep -q " BUILD_ONLY "; then
    run_builder
elif echo " $* " | grep -q " REPORT_ONLY "; then
    run_report
else
    run_builder
    run_report
    run_simer
fi
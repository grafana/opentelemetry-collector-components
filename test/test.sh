#!/bin/bash

while getopts d:s:b:g: flag
do
    case "${flag}" in
        d) distribution=${OPTARG};;
    esac
done
if [[ -z $distribution ]]; then
    echo "Distribution to test not provided. Use '-d' to specify the names of the distribution to test. Ex.:"
    echo "$0 -d tracing"
    exit 1
fi

# register the teardown function before we can use it in the trap
function teardown {
    ## tear down
    echo "🔧 Tearing down..."
    ./test/stop-otelcol.sh -d ${distribution}

    mkdir -p ./test/logs
    
    echo "🪵 '${distribution}' distribution of the OpenTelemetry Collector logs"
    cat ./test/logs/otelcol-${distribution}.log

    echo "🪵 Test logs"
    cat ./test/logs/test-${distribution}.log
}

## setup
echo "🔧 Setting up..."
for st in ./test/install-tracegen.sh
do
    ./${st}
    rc=$?
    if [ $rc != 0 ]; then
        exit $rc
    fi
done

# from this point and on, we run the teardown before we exit
trap teardown EXIT

## test
echo "🔧 Starting '${distribution}' distribution of the OpenTelemetry Collector..."
./test/start-otelcol.sh -d ${distribution}
rc=$?
if [ $rc != 0 ]; then
    exit $rc
fi

## generate a trace
echo "🔧 Generating trace..."
./test/generate-trace.sh -d ${distribution}
rc=$?
if [ $rc != 0 ]; then
    exit $rc
fi

## check that a trace was received
echo "🔧 Checking for existence of a trace..."
./test/check-trace.sh -d ${distribution}
rc=$?
if [ $rc != 0 ]; then
    exit $rc
fi

echo "✅ PASS: '${distribution}'"
exit 0

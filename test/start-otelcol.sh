#!/bin/bash

while getopts d:s:b:g: flag
do
    case "${flag}" in
        d) distribution=${OPTARG};;
    esac
done
if [[ -z $distribution ]]; then
    echo "Distributioon to test not provided. Use '-d' to specify the names of the distribution to test. Ex.:"
    echo "$0 -d tracing"
    exit 1
fi

max_retries=50

# start the distribution
./distributions/${distribution}/_build/otelcol --config test/config/${distribution}.yaml  > ./test/logs/otelcol-${distribution}.log 2>&1 &
pid=$!

retries=0
while true
do
    kill -0 "${pid}" >/dev/null 2>&1
    if [ $? != 0 ]; then
        echo "❌ FAIL. The Grafana Labs '${distribution}' distribution of the OpenTelemetry Collector isn't running. Startup log:"
        failed=true
        exit 1
    fi

    curl -s localhost:13133 | grep "Server available" > /dev/null
    if [ $? == 0 ]; then
        echo "✅ The Grafana Labs '${distribution}' distribution of the OpenTelemetry Collector started."
        echo "${pid}" > "otelcol-${distribution}.pid"
        break
    fi

    echo "Server still unavailable" >> ./test/logs/test-${distribution}.log

    let "retries++"
    if [ "$retries" -gt "$max_retries" ]; then
        echo "❌ FAIL. Server wasn't up after about 5s."

        kill "${pid}"
        if [ $? != 0 ]; then
            echo "Failed to stop the running instance. Return code: $? . Skipping tests."
            exit 8
        fi
        exit 16
    fi
    sleep 0.1s
done

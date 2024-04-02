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

pid=$(cat otelcol-${distribution}.pid)
if [[ -z $pid ]]; then
    echo "No Collectors running. Nothing to stop."
    exit 0
fi

kill "${pid}"
if [ $? != 0 ]; then
    echo "Failed to stop the running instance. Return code: $? . Skipping tests."
    exit 2
fi

while kill -0 "${pid}" >/dev/null 2>&1
do
    sleep 0.1s
done

rm "otelcol-${distribution}.pid"
echo "✅ '${distribution}' distribution of the OpenTelemetry Collector stopped."
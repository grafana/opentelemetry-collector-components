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

pid=$(cat otelcol-${distribution}.pid)
kill "${pid}"
if [ $? != 0 ]; then
    echo "Failed to stop the running instance. Return code: $? . Skipping tests."
    exit 2
fi

while kill -0 ${pid}
do
    sleep 0.1s
done

echo "✅ Grafana Labs '${distribution}' distribution of the OpenTelemetry Collector stopped."
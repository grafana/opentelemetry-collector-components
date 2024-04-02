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

tracegen -otlp-endpoint localhost:4317 -otlp-insecure -service e2e-test &>> ./test/logs/tracegen-${distribution}.log
if [ $? != 0 ]; then
    echo "Failed to generate a trace."
    exit 1
fi

echo "✅ Traces generated."

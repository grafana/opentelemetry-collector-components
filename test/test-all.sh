#!/bin/bash

while getopts d:s:b:g: flag
do
    case "${flag}" in
        d) distributions=${OPTARG};;
    esac
done

if [[ -z $distributions ]]; then
    echo "List of distributions to test not provided. Use '-d' to specify the names of the distributions to test. Ex.:"
    echo "$0 -d sidecar,tracing"
    exit 1
fi

echo "Distributions to test: $distributions";

for distribution in $(echo "$distributions" | tr "," "\n")
do
    ./test/test.sh -d "${distribution}"
    rc=$?
    if [ $rc != 0 ]; then
        echo "❌ FAIL. Test failed for '${distribution}' distribution."
        exit $rc
    fi
done

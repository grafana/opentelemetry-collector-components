#!/usr/bin/env bash
set -eufo pipefail

drone jsonnet --source .drone/drone.jsonnet --target .drone/drone.yml --stream --format
drone lint .drone/drone.yml
drone sign --save grafana/opentelemetry-collector-components .drone/drone.yml

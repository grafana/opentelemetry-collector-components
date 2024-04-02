# OpenTelemetry Collector distributions

This repository has private collection of OpenTelemetry Collector distributions used by Grafana Labs in non-supported settings.

No support is provided by distributions present in this repository.

## Adding a new distribution

To add a new distribution to this repository:

1) create a directory under `distributions` and place the `manifest.yaml` there
2) add `./github/workflows/ci-<distribution>.yaml` and `./github/workflows/release-<distribution>.yaml` files based on one of the existing distributions

You can test your new distribution with:

```console
./test/test.sh -d YOUR_DISTRIBUTION
```

Or, to run everything the CI would run:

```console
make ci
```

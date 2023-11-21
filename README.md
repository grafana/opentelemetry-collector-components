# Deprecation notice

This repository is now archived. The distributions we had here are now [available at a new home](https://github.com/jpkrohling/otelcol-distributions).

# OpenTelemetry Collector components by Grafana Labs

This repository will contain a set of components created by Grafana Labs and not available upstream ([open-telemetry/opentelemetry-collector](https://github.com/open-telemetry/opentelemetry-collector) or [open-telemetry/opentelemetry-collector-contrib](https://github.com/open-telemetry/opentelemetry-collector-contrib/)), typically because the components are experimental.

You will also find our custom distributions, mostly suitable for usage with Grafana Cloud.

We prefer to contribute upstream, which is why we see this repository here as only a temporary home for most components.

## Adding a new distribution

To add a new distribution to this repository:

1) create a directory under `distributions` and place the `manifest.yaml` there
2) change the `Makefile`'s `DISTRIBUTIONS` var to include the new distribution
3) add a configuration file in the `test/config` with your distribution's name

You can test your new distribution with:

```console
./test/test.sh -d YOUR_DISTRIBUTION
```

Or, to run everything the CI would run:

```console
make test
```

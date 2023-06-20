# Multiexporter Exporter

| Status                   |                       |
|--------------------------|-----------------------|
| Stability                | [alpha]               |
| Supported pipeline types | traces, metrics, logs |
| Distributions            | [grafana]             |

Creates exporters with endpoints that correspond to mappings values and
provided configuration.

## Configuration

- `mapping_key`: the key is used to lookup the mapping value from the context.
- `metrics/traces/logs`:
  - `exporter`: the exporter type. Currently, only `loki`, `otlp`, and `otlphttp` are supported.
  - `mapping`: an arbitrary key to the endpoint URL mapping.
  - `http_client` or `grpc_client`: See the [gRPC client config](https://github.com/open-telemetry/opentelemetry-collector/blob/58628a7820a1a2e5ba434e4b7ea629223165332a/config/configgrpc/README.md#client-configuration)
     and the [http client config](https://github.com/open-telemetry/opentelemetry-collector/tree/58628a7820a1a2e5ba434e4b7ea629223165332a/config/confighttp#client-configuration)
  - for configuring `retry`, `timeout`, `queue` settings see the
    [exporterhelper configuration](https://github.com/open-telemetry/opentelemetry-collector/tree/25129b794d488cb5bff4cf9ba48f604cdc4b03f1/exporter/exporterhelper#configuration)

An exporter will be created for each mapping entry. The exporters for the same signal
share the provided configuration. 

Example of usage:

```yaml
receivers:
  otlp:
    protocols:
      grpc:
        include_metadata: true
      http:
        include_metadata: true

exporters:
  multiexporter:
    mapping_key: 'X-Scope-InstanceURL'
    metrics:
      exporter: otlphttp
      http_client:
        tls:
          insecure: true
        auth:
          authenticator: headers_setter
      mapping:
        "https://prometheus-dev-02-dev-us-central-0.grafana.net": "..."
        "https://prometheus-dev-01-dev-us-central-0.grafana.net": "..."
    logs:
      exporter: loki
      http_client:
        tls:
          insecure: true
        auth:
          authenticator: headers_setter
      mapping:
        "https://tempo-dev-01-dev-us-central-0.grafana.net": "..."
    traces:
      exporter: otlp
      grpc_client:
        tls:
          insecure: true
        auth:
          authenticator: headers_setter
      mapping:
        "https://prometheus-dev-01-dev-us-central-0.grafana.net": "..."

processors:
  gcomapi:
    client:
      endpoint: "http://fake:3000"
      key: "fake"

service:
  pipelines:
    traces:
      receivers: [ otlp ]
      processors: [ gcomapi ]
      exporters: [ multiexporter ]
    metrics:
      receivers: [ otlp ]
      processors: [ gcomapi ]
      exporters: [ multiexporter ]
    logs:
      receivers: [ otlp ]
      processors: [ gcomapi ]
      exporters: [ multiexporter ]
```
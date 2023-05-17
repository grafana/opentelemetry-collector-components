# Debugging

To debug the distribution, use a configuration like this:

```yaml
    exporters:
        loki:
            endpoint: http://localhost:8080/loki/api/v1/push
            tls:
                insecure: true
                insecure_skip_verify: true
        otlp:
            endpoint: localhost:8081
            tls:
                insecure: false
                insecure_skip_verify: true
        otlphttp:
            endpoint: http://localhost:8082
            tls:
                insecure: true
                insecure_skip_verify: true
    extensions:
        health_check: {}
    processors:
        gcomapi:
            client:
                endpoint: "https://grafana-dev.com/api"
                key: "..."
        routing:
            attribute_source: context
            from_attribute: X-Scope-InstanceURL
            table:
                - exporters:
                    - loki
                  value: https://logs-dev-005.grafana-dev.net
                - exporters:
                    - otlphttp
                  value: https://prometheus-dev-01-dev-us-central-0.grafana-dev.net
                - exporters:
                    - otlp
                  value: https://tempo-dev-01-dev-us-central-0.grafana.net
    receivers:
        otlp:
            protocols:
                http:
                    include_metadata: "true"
    service:
        extensions:
            - health_check
        pipelines:
            logs:
                exporters:
                    - loki
                processors:
                    - gcomapi
                    - routing
                receivers:
                    - otlp
            metrics:
                exporters:
                    - otlphttp
                processors:
                    - gcomapi
                    - routing
                receivers:
                    - otlp
            traces:
                exporters:
                    - otlp
                processors:
                    - gcomapi
                    - routing
                receivers:
                    - otlp
```

The key can be created via https://grafana-dev.com/orgs/raintank/api-keys. Make sure it's an API with the `Admin` role.

With that, create port-forwards for the internal backends:

```
kubectl port-forward -n loki-dev-005  svc/cortex-gw-internal 8080:80
kubectl port-forward -n tempo-dev-01  svc/cortex-gw-internal 8081:80
kubectl port-forward -n cortex-dev-01 svc/cortex-gw-internal 8082:80
```

Finally, generate a `go.work` file, if you haven't yet:

```
go work init
go work use -r .
```

You might also want to get the vendor sources, to make it easier to debug:

```
cd distributions/otel-grafana/_build/
go mod vendor
```

With that in place, you can now run the distribution in debug mode (start point: `distributions/otel-grafana/_build/main.go`). If you are using VS Code, a `launch.json` similar to the following can be used:

```json
{
    // Use IntelliSense to learn about possible attributes.
    // Hover to view descriptions of existing attributes.
    // For more information, visit: https://go.microsoft.com/fwlink/?linkid=830387
    "version": "0.2.0",
    "configurations": [
        {
            "name": "Launch Package",
            "type": "go",
            "request": "launch",
            "mode": "auto",
            "program": "${fileDirname}",
            "args": ["--config", "/home/jpkroehling/Projects/github.com/jpkrohling/otelcol-configs/dev-us-central-0-v2.yaml"]
        }
    ]
}
```
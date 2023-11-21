FROM alpine:3

ARG USER_UID=10001
USER ${USER_UID}

COPY otel-grafana /otel-grafana
COPY configs/otel-grafana.yaml /etc/otel-grafana/config.yaml
ENTRYPOINT ["/otel-grafana"]
CMD ["--config", "/etc/otel-grafana/config.yaml"]
EXPOSE 4317

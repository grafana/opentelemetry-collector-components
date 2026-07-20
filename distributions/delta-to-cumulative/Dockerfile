FROM alpine:latest@sha256:28bd5fe8b56d1bd048e5babf5b10710ebe0bae67db86916198a6eec434943f8b

ARG USER_UID=10001
USER ${USER_UID}

COPY delta-to-cumulative /delta-to-cumulative
COPY otelcol.yaml /etc/delta-to-cumulative/config.yaml
ENTRYPOINT ["/delta-to-cumulative"]
CMD ["--config", "/etc/delta-to-cumulative/config.yaml"]
EXPOSE 4317

FROM alpine:latest

ARG USER_UID=10001
USER ${USER_UID}

COPY delta-to-cumulative /delta-to-cumulative
COPY otelcol.yaml /etc/delta-to-cumulative/config.yaml
ENTRYPOINT ["/delta-to-cumulative"]
CMD ["--config", "/etc/delta-to-cumulative/config.yaml"]
EXPOSE 4317

FROM alpine:latest@sha256:5b10f432ef3da1b8d4c7eb6c487f2f5a8f096bc91145e68878dd4a5019afde11

ARG USER_UID=10001
USER ${USER_UID}

COPY delta-to-cumulative /delta-to-cumulative
COPY otelcol.yaml /etc/delta-to-cumulative/config.yaml
ENTRYPOINT ["/delta-to-cumulative"]
CMD ["--config", "/etc/delta-to-cumulative/config.yaml"]
EXPOSE 4317

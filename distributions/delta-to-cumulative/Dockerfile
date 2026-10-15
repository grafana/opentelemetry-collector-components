FROM alpine:latest@sha256:294b683cb724975bec92580e1e685676bd4b50bda910ddb8c51d4cabeaec77e6

ARG USER_UID=10001
USER ${USER_UID}

COPY delta-to-cumulative /delta-to-cumulative
COPY otelcol.yaml /etc/delta-to-cumulative/config.yaml
ENTRYPOINT ["/delta-to-cumulative"]
CMD ["--config", "/etc/delta-to-cumulative/config.yaml"]
EXPOSE 4317

FROM alpine:3.16

ARG USER_UID=10001
USER ${USER_UID}

COPY sidecar /sidecar
COPY configs/sidecar.yaml /etc/sidecar/config.yaml
ENTRYPOINT ["/sidecar"]
CMD ["--config", "/etc/sidecar/config.yaml"]
EXPOSE 4317

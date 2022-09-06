FROM alpine:3.16

ARG USER_UID=10001
USER ${USER_UID}

COPY tracing /tracing
COPY configs/tracing.yaml /etc/tracing/config.yaml
ENTRYPOINT ["/tracing"]
CMD ["--config", "/etc/tracing/config.yaml"]
EXPOSE 4317

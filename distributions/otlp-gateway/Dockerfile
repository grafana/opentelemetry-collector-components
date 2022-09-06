FROM alpine:3.16

ARG USER_UID=10001
USER ${USER_UID}

COPY otlp-gateway /otlp-gateway
COPY configs/otlp-gateway.yaml /etc/otlp-gateway/config.yaml
ENTRYPOINT ["/otlp-gateway"]
CMD ["--config", "/etc/otlp-gateway/config.yaml"]
EXPOSE 4317

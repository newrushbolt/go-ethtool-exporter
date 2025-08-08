# For test purpose only
FROM alpine:3.18
RUN apk add --no-cache ethtool iproute2

ADD go-ethtool-exporter /go-ethtool-exporter

ENTRYPOINT ["/go-ethtool-exporter"]
CMD ["--discover-all-ports", "--collect-all-metrics", "http-server"]

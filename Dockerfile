
FROM golang:1.22 as builder

ENV GO111MODULE=on
ENV CGO_ENABLED=0

COPY / /work
WORKDIR /work

RUN make latency-exporter

FROM scratch
COPY --from=builder /work/bin/latency-exporter /latency-exporter

USER 999
ENTRYPOINT ["/latency-exporter"]

EXPOSE 9080

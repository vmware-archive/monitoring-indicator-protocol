FROM golang:1.13 as builder

WORKDIR /root

ENV GOOS=linux \
    GOARCH=amd64 \
    CGO_ENABLED=0

COPY /go.mod /go.sum /root/

RUN go version && \
    go mod download

COPY / /root/

RUN go build \
    -a \
    -installsuffix nocgo \
    -o /prometheus-controller \
    -mod=readonly \
    cmd/prometheus_rules_controller/main.go

FROM ubuntu

COPY --from=builder /prometheus-controller /srv/
COPY cmd/prometheus_rules_controller/docker_run.sh /srv/run.sh
WORKDIR /srv
CMD ["./run.sh"]

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
    -o /indicator-registry-proxy \
    -mod=readonly \
    cmd/registry_proxy/main.go

FROM ubuntu

COPY --from=builder /indicator-registry-proxy /srv/
COPY cmd/registry_proxy/docker_run.sh /srv/run.sh
WORKDIR /srv
CMD ["./run.sh"]

FROM golang:1.12 as builder

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
    -o /indicator-registry-agent \
    -mod=readonly \
    cmd/registry_agent/main.go

FROM ubuntu

COPY --from=builder /indicator-registry-agent /srv/
WORKDIR /srv
CMD [ "./indicator-registry-agent", "--tls-pem-path", "certs/client.pem", "--tls-key-path", "certs/client.key", "--tls-root-ca-pem", "certs/ca.pem", "--tls-server-cn", "indicator-registry-proxy", "--registry", "https://indicator-registry-proxy:10567", "--documents-glob", "./resources/indicators.yml"]

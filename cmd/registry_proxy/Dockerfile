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
WORKDIR /srv
CMD [ "./indicator-registry-proxy", \
  "--tls-pem-path", "certs/server.pem", \
  "--tls-key-path", "certs/server.key", \
  "--tls-root-ca-pem", "certs/ca.pem", \
  "--tls-server-cn", "localhost", \
  "--tls-client-pem-path", "certs/client.pem", \
  "--tls-client-key-path", "certs/client.key", \
  "--local-registry-addr", "indicator-registry:10568" ]

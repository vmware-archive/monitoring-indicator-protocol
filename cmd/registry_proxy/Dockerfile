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
    -o /indicator-registry-proxy \
    -mod=readonly \
    cmd/registry_proxy/main.go

FROM scratch
# ^ should this be scratch or debian? (same question for the registry dockerfile)
# I think scratch once we're done debugging it, though since it's a dev image there's maybe something to be said for an os with bash and a package manager
# Ubuntu may be the better choice over Debian though since Pivotal works with Canonical a lot?

COPY --from=builder /indicator-registry-proxy /srv/
WORKDIR /srv
CMD [ "./indicator-registry-proxy", "--tls-pem-path", "certs/server.pem", "--tls-key-path", "certs/server.key", "--tls-root-ca-pem", "certs/ca.pem", "--tls-server-cn", "localhost", "--tls-client-pem-path", "certs/client.pem", "--tls-client-key-path", "certs/client.key", "--local-registry-addr", "indicator-registry:10568" ]

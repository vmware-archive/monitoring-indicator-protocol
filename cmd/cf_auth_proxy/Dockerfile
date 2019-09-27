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
    -o /cf-auth-proxy \
    -mod=readonly \
    cmd/cf_auth_proxy/main.go

FROM ubuntu

COPY --from=builder /cf-auth-proxy /srv/
COPY cmd/cf_auth_proxy/docker_run.sh /srv/run.sh
WORKDIR /srv
CMD ["./run.sh"]

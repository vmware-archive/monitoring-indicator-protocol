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
    -o /indicator-registry \
    -mod=readonly \
    cmd/registry/main.go

FROM ubuntu
RUN ["apt-get", "update"]
RUN ["apt-get", "install", "git-core", "-y"]

COPY --from=builder /indicator-registry /srv/
WORKDIR /srv
CMD [ "./indicator-registry", \
  "--host", "", \
  "--config", "./resources/config.yml"]

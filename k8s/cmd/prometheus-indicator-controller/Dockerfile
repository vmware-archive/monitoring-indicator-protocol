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
    -o /prometheus-indicator-controller \
    -mod=readonly \
    k8s/cmd/prometheus-indicator-controller/main.go

FROM scratch

COPY --from=builder /prometheus-indicator-controller /srv/
WORKDIR /srv
CMD [ "/srv/prometheus-indicator-controller" ]

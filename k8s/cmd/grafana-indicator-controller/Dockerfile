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
    -o /grafana-indicator-controller \
    -mod=readonly \
    k8s/cmd/grafana-indicator-controller/main.go

FROM scratch

COPY --from=builder /grafana-indicator-controller /srv/
WORKDIR /srv
CMD [ "/srv/grafana-indicator-controller" ]

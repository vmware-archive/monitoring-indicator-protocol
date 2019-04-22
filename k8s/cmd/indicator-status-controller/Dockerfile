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
    -o /indicator-status-controller \
    -mod=readonly \
    k8s/cmd/indicator-status-controller/main.go

FROM scratch

COPY --from=builder /indicator-status-controller /srv/
WORKDIR /srv
CMD [ "/srv/indicator-status-controller" ]

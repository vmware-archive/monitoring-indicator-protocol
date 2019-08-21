#!/usr/bin/env bash
set -efu

export CERTS=./test_fixtures
export REGISTRY_PORT=8092
export REGISTRY_PROXY_PORT=8091
export REGISTRY_URL="https://localhost:${REGISTRY_PROXY_PORT}"
export INDICATOR_DOCUMENTS='./example_indicators*.yml'

echo "Starting registry on PORT $REGISTRY_PORT"
go run cmd/registry/main.go \
  -port ${REGISTRY_PORT} \
  -indicator-expiration 1m \
    2>&1 | sed 's/^/REGISTRY: /' \
  &
sleep 3

echo "Starting registry proxy on PORT $REGISTRY_PROXY_PORT"
go run cmd/registry_proxy/main.go \
  -port ${REGISTRY_PROXY_PORT} \
  -host localhost \
  -tls-key-path ${CERTS}/server.key \
  -tls-pem-path ${CERTS}/server.pem \
  -tls-client-key-path ${CERTS}/client.key \
  -tls-client-pem-path ${CERTS}/client.pem \
  -tls-root-ca-pem ${CERTS}/ca.pem \
  -local-registry-addr localhost:${REGISTRY_PORT} \
    2>&1 | sed 's/^/REGISTRY PROXY: /' \
  &
sleep 3

echo "Starting registry agent for $REGISTRY_URL"
go run cmd/registry_agent/main.go \
  -interval 5s \
  -tls-key-path ${CERTS}/client.key \
  -tls-pem-path ${CERTS}/client.pem \
  -tls-root-ca-pem ${CERTS}/ca.pem \
  -tls-server-cn localhost \
  -registry ${REGISTRY_URL} \
  -documents-glob ${INDICATOR_DOCUMENTS} \
    2>&1 | sed 's/^/REGISTRY AGENT: /' \
  &

wait

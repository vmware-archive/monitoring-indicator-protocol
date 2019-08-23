#!/usr/bin/env bash
set -efu

export CERTS=./test_fixtures
export REGISTRY_PROXY_PORT=8091
export REGISTRY_URL="https://localhost:${REGISTRY_PROXY_PORT}"

echo "Starting auth proxy on PORT 5050"
go run cmd/cf_auth_proxy/main.go \
  -port 5050 \
  -host localhost \
  -tls-key-path ${CERTS}/server.key \
  -tls-pem-path ${CERTS}/server.pem \
  -tls-client-key-path ${CERTS}/client.key \
  -tls-client-pem-path ${CERTS}/client.pem \
  -tls-root-ca-pem ${CERTS}/ca.pem \
  -uaa-addr "https://uaa.madlamp.cf-denver.com" \
  -registry-addr ${REGISTRY_URL} \
    2>&1 | sed 's/^/cf auth PROXY: /' \
  &

wait

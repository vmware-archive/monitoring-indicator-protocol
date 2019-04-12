#!/usr/bin/env bash
set -efu

export CERTS=./test_fixtures
export REGISTRY_HOST="https://localhost:8091"

echo "Running prometheus agent"
go run cmd/prometheus_rules_controller/main.go \
  -registry ${REGISTRY_HOST} \
  -output-directory /tmp/alerts \
  -tls-pem-path ${CERTS}/client.pem \
  -tls-key-path ${CERTS}/client.key \
  -tls-root-ca-pem ${CERTS}/ca.pem \
  -tls-server-cn localhost \
  -prometheus http://localhost:9090


#!/usr/bin/env bash
set -efu

ENVIRONMENT=sunstorm

set +u
  source ~/workspace/denver-bash-it/custom/environment-targeting.bash
  target "$ENVIRONMENT"
set -u

export CERTS=./test_fixtures
export REGISTRY_URI="https://localhost:8091"
export PROMETHEUS_URI="https://log-cache.${ENVIRONMENT}.cf-denver.com "
export UAA_URI="https://login.${ENVIRONMENT}.cf-denver.com"
export UAA_CLIENT_ID="firehose_exporter"
export UAA_CLIENT_SECRET=$(credhub g -n /bosh-${ENVIRONMENT}/cf-01234567890123456789/uaa_clients_firehose_exporter_secret -j | jq -r .value)
export INTERVAL=1m

echo "Running status controller"
go run cmd/status_controller/main.go \
  -registry-uri ${REGISTRY_URI} \
  -prometheus-uri ${PROMETHEUS_URI} \
  -interval ${INTERVAL} \
  -oauth-server ${UAA_URI} \
  -oauth-client-id ${UAA_CLIENT_ID} \
  -oauth-client-secret ${UAA_CLIENT_SECRET} \
  -tls-pem-path ${CERTS}/client.pem \
  -tls-key-path ${CERTS}/client.key \
  -tls-root-ca-pem ${CERTS}/server.pem \
  -tls-server-cn localhost 


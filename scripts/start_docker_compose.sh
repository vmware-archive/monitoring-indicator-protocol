#!/usr/bin/env bash

set +u
  source ~/workspace/denver-bash-it/custom/environment-targeting.bash
  target madlamp
set -u

pushd ~/workspace/monitoring-indicator-protocol/test_fixtures || exit
  export TLS_PEM=$(cat client.pem)
  export TLS_KEY=$(cat client.key)
  export CLIENT_PEM=$(cat client.pem)
  export CLIENT_KEY=$(cat client.key)
  export SERVER_PEM=$(cat server.pem)
  export SERVER_KEY=$(cat server.key)
  export TLS_ROOT_CA_PEM=$(cat ca.pem)
  export UAA_ADDRESS=https://uaa.madlamp.cf-denver.com
  export PROMETHEUS_URI="https://metric-store.madlamp.cf-denver.com"
  export UAA_URI=$UAA_ADDRESS
  export UAA_CLIENT_ID="apps_metrics_processing"
  export UAA_CLIENT_SECRET=$(credhub g -n /bosh-madlamp/cf-01234567890123456789/uaa_clients_cc-service-dashboards_secret -j | jq -r .value)
popd || exit


pushd docker-compose || exit
  docker-compose up
popd || exit

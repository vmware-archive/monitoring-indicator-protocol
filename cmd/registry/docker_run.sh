#!/usr/bin/env bash
# Entrypoint for the dockerfile
set -e

if [[ ! -d certs ]]; then
    mkdir certs
fi

if [[ ! -z "$TLS_PEM" ]]; then
    echo "$TLS_PEM" > certs/client.pem
fi

if [[ ! -z "$TLS_KEY" ]]; then
    echo "$TLS_KEY" > certs/client.key
fi

if [[ ! -z "$TLS_ROOT_CA_PEM" ]]; then
    echo "$TLS_ROOT_CA_PEM" > certs/ca.pem
fi

./indicator-registry-agent \
  --tls-pem-path certs/client.pem \
  --tls-key-path certs/client.key \
  --tls-root-ca-pem certs/ca.pem \
  --tls-server-cn indicator-registry-proxy \
  --registry https://indicator-registry-proxy:10567 \
  --documents-glob ./resources/indicators.yml \
  --interval 5s

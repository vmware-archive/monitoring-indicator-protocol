#!/usr/bin/env bash
# Entrypoint for the dockerfile
set -e

: "${TLS_PEM:?}"
: "${TLS_KEY:?}"
: "${TLS_ROOT_CA_PEM:?}"
: "${PROMETHEUS_URI:?}"
: "${UAA_CLIENT_ID:?}"
: "${UAA_URI:?}"
: "${UAA_CLIENT_SECRET:?}"

if [ ! -d "certs" ]
then
  mkdir certs
fi

echo -e "$TLS_PEM" > certs/client.pem
echo -e "$TLS_KEY" > certs/client.key
echo -e "$TLS_ROOT_CA_PEM" > certs/ca.pem

./status-controller \
  -registry-uri https://indicator-registry-proxy:10567 \
  -prometheus-uri ${PROMETHEUS_URI} \
  -oauth-server ${UAA_URI} \
  -oauth-client-id ${UAA_CLIENT_ID} \
  -oauth-client-secret ${UAA_CLIENT_SECRET} \
  -tls-pem-path certs/client.pem \
  -tls-key-path certs/client.key \
  -tls-root-ca-pem certs/ca.pem \
  -tls-server-cn localhost \
  -interval 30s \
  -k

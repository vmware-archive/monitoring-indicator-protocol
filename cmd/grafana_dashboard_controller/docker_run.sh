#!/usr/bin/env bash
# Entrypoint for the dockerfile
set -e

: "${TLS_PEM:?}"
: "${TLS_KEY:?}"
: "${TLS_ROOT_CA_PEM:?}"

if [ ! -d "certs" ]
then
  mkdir certs
fi

if [ ! -d "/srv/dashboards" ]
then
  mkdir /srv/dashboards
fi

echo -e "$TLS_PEM" > certs/client.pem
echo -e "$TLS_KEY" > certs/client.key
echo -e "$TLS_ROOT_CA_PEM" > certs/ca.pem

./grafana-controller \
  -registry https://indicator-registry-proxy:10567 \
  -output-directory /srv/dashboards \
  -tls-pem-path certs/client.pem \
  -tls-key-path certs/client.key \
  -tls-root-ca-pem certs/ca.pem \
  -tls-server-cn localhost

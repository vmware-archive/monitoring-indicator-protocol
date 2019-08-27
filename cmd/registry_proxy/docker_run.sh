#!/usr/bin/env bash
# Entrypoint for the dockerfile
set -e

: "${CLIENT_PEM:?}"
: "${CLIENT_KEY:?}"
: "${SERVER_PEM:?}"
: "${SERVER_KEY:?}"
: "${TLS_ROOT_CA_PEM:?}"

if [ ! -d "certs" ]
then
  mkdir certs
fi

echo -e "$CLIENT_PEM" > certs/client.pem
echo -e "$CLIENT_KEY" > certs/client.key
echo -e "$SERVER_PEM" > certs/server.pem
echo -e "$SERVER_KEY" > certs/server.key
echo -e "$TLS_ROOT_CA_PEM" > certs/ca.pem

./indicator-registry-proxy \
  --tls-pem-path certs/server.pem \
  --tls-key-path certs/server.key \
  --tls-root-ca-pem certs/ca.pem \
  --tls-server-cn localhost \
  --tls-client-pem-path certs/client.pem \
  --tls-client-key-path certs/client.key \
  --local-registry-addr indicator-registry:10568

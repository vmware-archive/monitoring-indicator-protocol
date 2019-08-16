#!/usr/bin/env bash
# Entrypoint for the dockerfile
set -e

if [[ ! -d certs ]]; then
    mkdir certs
fi

if [[ ! -z "$CLIENT_PEM" ]]; then
    echo -e "$CLIENT_PEM" > certs/client.pem
fi

if [[ ! -z "$CLIENT_KEY" ]]; then
    echo -e "$CLIENT_KEY" > certs/client.key
fi

if [[ ! -z "$SERVER_PEM" ]]; then
    echo -e "$SERVER_PEM" > certs/server.pem
fi

if [[ ! -z "$SERVER_KEY" ]]; then
    echo -e "$SERVER_KEY" > certs/server.key
fi

if [[ ! -z "$TLS_ROOT_CA_PEM" ]]; then
    echo -e "$TLS_ROOT_CA_PEM" > certs/ca.pem
fi


./indicator-registry-proxy \
  --tls-pem-path certs/server.pem \
  --tls-key-path certs/server.key \
  --tls-root-ca-pem certs/ca.pem \
  --tls-server-cn localhost \
  --tls-client-pem-path certs/client.pem \
  --tls-client-key-path certs/client.key \
  --local-registry-addr indicator-registry:10568

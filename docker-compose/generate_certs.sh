#!/usr/bin/env bash

openssl genrsa -out certs/ca.key 4096
openssl req -x509 -new -nodes -key certs/ca.key -subj "/CN=localhost" -sha256 -days 1024 -out certs/ca.pem
openssl genrsa -out certs/server.key 2048
openssl req -new -sha256 -key certs/server.key -subj "/CN=indicator-registry-proxy" -out certs/server.csr
openssl x509 -req -in certs/server.csr -CA certs/ca.pem -CAkey certs/ca.key -CAcreateserial -out certs/server.pem -days 500 -sha256
openssl genrsa -out certs/client.key 2048
openssl req -new -sha256 -key certs/client.key -subj "/CN=localhost" -out certs/client.csr
openssl x509 -req -in certs/client.csr -CA certs/ca.pem -CAkey certs/ca.key -CAcreateserial -out certs/client.pem -days 500 -sha256

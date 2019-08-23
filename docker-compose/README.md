## Developing with Docker
We have provided a script to create the necessary certificates and start your docker container for you. If you have the repo cloned, run `./scripts/start_docker_compose.sh` from the root. The registry will be running on port 10567 by default. To curl this registry, reference the certs created in the certs directory within docker-compose. For example:

```bash
curl https://localhost:10567/v1/indicator-documents -k --key test_fixtures/client.key --cert test_fixtures/client.pem --cacert test_fixtures/ca.key
```

Any indicator document, patch or config files you're working with need to be added to docker-compose/resources. The images are mounted with config.yml and indicators.yml specifically.

## Rebuilding Docker Images
If you want to rebuild the docker images to reflect local changes, run the following commands.

```bash
docker build -t indicatorprotocol/bosh-indicator-protocol-registry -f cmd/registry/Dockerfile .
docker build -t indicatorprotocol/bosh-indicator-protocol-registry-proxy -f cmd/registry_proxy/Dockerfile .
docker build -t indicatorprotocol/bosh-indicator-protocol-registry-agent -f cmd/registry_agent/Dockerfile .
docker build -t indicatorprotocol/bosh-indicator-protocol-cf-auth-proxy -f cmd/cf_auth_proxy/Dockerfile .
docker build -t indicatorprotocol/bosh-indicator-protocol-status-controller -f cmd/status_controller/Dockerfile .
docker build -t indicatorprotocol/bosh-indicator-protocol-prometheus-controller -f cmd/prometheus_rules_controller/Dockerfile .
docker build -t indicatorprotocol/bosh-indicator-protocol-grafana-controller -f cmd/grafana_dashboard_controller/Dockerfile .
```

## Certs
You need to define the following cert environment variables in your shell:
- TLS_PEM: registration agent client cert
- TLS_KEY: registration agent client key
- CLIENT_PEM: registry proxy client cert
- CLIENT_KEY: registry proxy client key
- SERVER_PEM: registry proxy server cert
- SERVER_KEY: registry proxy server key
- TLS_ROOT_CA_PEM: root CA for all certificates

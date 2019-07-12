## Developing with Docker
We have provided a script to create the necessary certificates and start your docker container for you. If you have the repo cloned, run `./scripts/start_docker_compose.sh` from the root. The registry will be running on port 10567 by default. To curl this registry, reference the certs created in the certs directory within docker-compose. For example:

```bash
curl https://localhost:10567/v1/indicator-documents -k --key docker-compose/certs/client.key --cert docker-compose/certs/client.pem --cacert docker-compose/certs/ca.key
```

Any indicator document, patch or config files you're working with need to be added to docker-compose/resources. The images are mounted with config.yml and indicators.yml specifically.

## Rebuilding Docker Images
If you want to rebuild the docker images to reflect local changes, run the following commands.

To build the registry:
```bash
docker build -t indicatorprotocol/bosh-indicator-protocol-registry -f cmd/registry/Dockerfile .
```

To build the registry proxy:
```bash
docker build -t indicatorprotocol/bosh-indicator-protocol-registry-proxy -f cmd/registry_proxy/Dockerfile .
```

To build the registry agent:
```bash
docker build -t indicatorprotocol/bosh-indicator-protocol-registry-agent -f cmd/registry_agent/Dockerfile .
```

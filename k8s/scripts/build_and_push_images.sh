#!/bin/bash

set -eEuo pipefail

cd ..

docker build \
    -t indicatorprotocol/grafana-indicator-controller:latest \
    -f k8s/cmd/grafana-indicator-controller/Dockerfile \
    .

docker build \
    -t indicatorprotocol/prometheus-indicator-controller:latest \
    -f k8s/cmd/prometheus-indicator-controller/Dockerfile \
    .

docker push indicatorprotocol/grafana-indicator-controller:latest
docker push indicatorprotocol/prometheus-indicator-controller:latest
#!/bin/bash

set -eEuo pipefail

pushd ~/workspace/monitoring-indicator-protocol

    docker build \
        -t indicatorprotocol/grafana-indicator-controller:dev \
        -f k8s/cmd/grafana-indicator-controller/Dockerfile \
        .

    docker build \
        -t indicatorprotocol/prometheus-indicator-controller:dev \
        -f k8s/cmd/prometheus-indicator-controller/Dockerfile \
        .

    docker build \
        -t indicatorprotocol/indicator-lifecycle-controller:dev \
        -f k8s/cmd/indicator-lifecycle-controller/Dockerfile \
        .

    docker build \
        -t indicatorprotocol/indicator-admission:dev \
        -f k8s/cmd/indicator-admission/Dockerfile \
        .

    docker push indicatorprotocol/grafana-indicator-controller:dev
    docker push indicatorprotocol/prometheus-indicator-controller:dev
    docker push indicatorprotocol/indicator-lifecycle-controller:dev
    docker push indicatorprotocol/indicator-admission:dev

popd

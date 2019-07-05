#!/bin/bash

set -eEuo pipefail

pushd ~/workspace/monitoring-indicator-protocol > /dev/null
    docker build \
        -t indicatorprotocol/k8s-grafana-indicator-controller:dev \
        -f k8s/cmd/grafana-indicator-controller/Dockerfile \
        .

    docker build \
        -t indicatorprotocol/k8s-prometheus-indicator-controller:dev \
        -f k8s/cmd/prometheus-indicator-controller/Dockerfile \
        .

    docker build \
        -t indicatorprotocol/k8s-indicator-lifecycle-controller:dev \
        -f k8s/cmd/indicator-lifecycle-controller/Dockerfile \
        .

    docker build \
        -t indicatorprotocol/k8s-indicator-admission:dev \
        -f k8s/cmd/indicator-admission/Dockerfile \
        .

    docker build \
        -t indicatorprotocol/k8s-indicator-status-controller:dev \
        -f k8s/cmd/indicator-status-controller/Dockerfile \
        .

    docker push indicatorprotocol/k8s-grafana-indicator-controller:dev
    docker push indicatorprotocol/k8s-prometheus-indicator-controller:dev
    docker push indicatorprotocol/k8s-indicator-lifecycle-controller:dev
    docker push indicatorprotocol/k8s-indicator-admission:dev
    docker push indicatorprotocol/k8s-indicator-status-controller:dev


    grafana_digest="$(
        docker inspect indicatorprotocol/k8s-grafana-indicator-controller:dev \
            | jq .[0].RepoDigests[0] --raw-output
    )"
    prometheus_digest="$(
        docker inspect indicatorprotocol/k8s-prometheus-indicator-controller:dev \
            | jq .[0].RepoDigests[0] --raw-output
    )"
    indicator_lifecycle_digest="$(
        docker inspect indicatorprotocol/k8s-indicator-lifecycle-controller:dev \
            | jq .[0].RepoDigests[0] --raw-output
    )"
    indicator_admission_digest="$(
        docker inspect indicatorprotocol/k8s-indicator-admission:dev \
            | jq .[0].RepoDigests[0] --raw-output
    )"
    indicator_status_digest="$(
        docker inspect indicatorprotocol/k8s-indicator-status-controller:dev \
            | jq .[0].RepoDigests[0] --raw-output
    )"

    mkdir -p k8s/overlays/dev
    echo "
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
bases:
- ../../config
" > k8s/overlays/dev/kustomization.yaml

    pushd k8s/overlays/dev > /dev/null
        kustomize edit set image "$grafana_digest"
        kustomize edit set image "$prometheus_digest"
        kustomize edit set image "$indicator_lifecycle_digest"
        kustomize edit set image "$indicator_admission_digest"
        kustomize edit set image "$indicator_status_digest"
    popd > /dev/null

    kubectl apply -k k8s/overlays/dev
popd > /dev/null

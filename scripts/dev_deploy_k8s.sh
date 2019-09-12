#!/bin/bash

set -eEuo pipefail

SCRIPT=`realpath $0`
SCRIPTDIR=`dirname $SCRIPT`
REPOROOT=$SCRIPTDIR/..

pushd ~/workspace/monitoring-indicator-protocol > /dev/null
    docker build \
        -t indicatorprotocol/k8s-grafana-indicator-controller:dev \
        -f k8s/cmd/grafana-indicator-controller/Dockerfile \
        $REPOROOT &

    docker build \
        -t indicatorprotocol/k8s-prometheus-indicator-controller:dev \
        -f k8s/cmd/prometheus-indicator-controller/Dockerfile \
        $REPOROOT &

    docker build \
        -t indicatorprotocol/k8s-indicator-lifecycle-controller:dev \
        -f k8s/cmd/indicator-lifecycle-controller/Dockerfile \
        $REPOROOT &

    docker build \
        -t indicatorprotocol/k8s-indicator-admission:dev \
        -f k8s/cmd/indicator-admission/Dockerfile \
        $REPOROOT &

    docker build \
        -t indicatorprotocol/k8s-indicator-status-controller:dev \
        -f k8s/cmd/indicator-status-controller/Dockerfile \
        $REPOROOT &

    wait

    docker push indicatorprotocol/k8s-grafana-indicator-controller:dev &
    docker push indicatorprotocol/k8s-prometheus-indicator-controller:dev &
    docker push indicatorprotocol/k8s-indicator-lifecycle-controller:dev &
    docker push indicatorprotocol/k8s-indicator-admission:dev &
    docker push indicatorprotocol/k8s-indicator-status-controller:dev &

    wait


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
    cp -r k8s/config k8s/overlays/dev

    pushd $REPOROOT/k8s/overlays/dev > /dev/null
        kustomize edit set image "$grafana_digest"
        kustomize edit set image "$prometheus_digest"
        kustomize edit set image "$indicator_lifecycle_digest"
        kustomize edit set image "$indicator_admission_digest"
        kustomize edit set image "$indicator_status_digest"
    popd > /dev/null

    kubectl apply -k $REPOROOT/k8s/overlays/dev/config
popd > /dev/null

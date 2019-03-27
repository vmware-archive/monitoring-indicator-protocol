#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail

SCRIPT_ROOT=$(dirname ${BASH_SOURCE})/..
CODEGEN_PKG=${CODEGEN_PKG:-$(cd ${SCRIPT_ROOT}; ls -d -1 ./vendor/k8s.io/code-generator 2>/dev/null)}

mkdir -p "$GOPATH/src/github.com/pivotal/monitoring-indicator-protocol/k8s/pkg/"

rm -rf "$GOPATH/src/github.com/pivotal/monitoring-indicator-protocol/k8s/pkg/client"
rm -rf "$GOPATH/src/github.com/pivotal/monitoring-indicator-protocol/k8s/pkg/apis"

rsync -avzci --delete "$SCRIPT_ROOT/pkg/apis/" "$GOPATH/src/github.com/pivotal/monitoring-indicator-protocol/k8s/pkg/apis/"

${CODEGEN_PKG}/generate-groups.sh "deepcopy,client,informer,lister" \
  github.com/pivotal/monitoring-indicator-protocol/k8s/pkg/client github.com/pivotal/monitoring-indicator-protocol/k8s/pkg/apis \
  "indicatordocument:v1alpha1"

rsync -avzci --delete "$GOPATH/src/github.com/pivotal/monitoring-indicator-protocol/k8s/pkg/client/" "$SCRIPT_ROOT/pkg/client/"
rsync -avzci --delete "$GOPATH/src/github.com/pivotal/monitoring-indicator-protocol/k8s/pkg/apis/" "$SCRIPT_ROOT/pkg/apis/"

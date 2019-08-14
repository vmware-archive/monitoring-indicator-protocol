#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail

REPO_ROOT="$(dirname $BASH_SOURCE)/.."
CODEGEN_DIR="$REPO_ROOT/vendor/k8s.io/code-generator"

rm -rf output_base
mkdir -p output_base

chmod +x "$CODEGEN_DIR/generate-groups.sh"
"$CODEGEN_DIR/generate-groups.sh" \
  all \
  github.com/pivotal/monitoring-indicator-protocol/pkg/k8s/client \
  github.com/pivotal/monitoring-indicator-protocol/pkg/k8s/apis \
  indicatordocument:v1 \
  --output-base "$PWD/output_base" \
  --go-header-file $REPO_ROOT/hack/boilerplate.go.txt
chmod -x "$CODEGEN_DIR/generate-groups.sh"

rsync -av "output_base/github.com/pivotal/monitoring-indicator-protocol/pkg/k8s/apis/" "$REPO_ROOT/pkg/k8s/apis/"
rsync -av --delete "output_base/github.com/pivotal/monitoring-indicator-protocol/pkg/k8s/client/" "$REPO_ROOT/pkg/k8s/client/"
rm -rf output_base

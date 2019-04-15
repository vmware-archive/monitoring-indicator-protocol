#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail

REPO_ROOT="$(dirname $BASH_SOURCE)/../.."
CODEGEN_DIR="$REPO_ROOT/vendor/k8s.io/code-generator"

rm -rf output_base
mkdir -p output_base

chmod +x "$CODEGEN_DIR/generate-groups.sh"
"$CODEGEN_DIR/generate-groups.sh" \
  all \
  github.com/pivotal/monitoring-indicator-protocol/k8s/pkg/client \
  github.com/pivotal/monitoring-indicator-protocol/k8s/pkg/apis \
  indicatordocument:v1alpha1 \
  --output-base "$PWD/output_base" \
  --go-header-file $REPO_ROOT/k8s/hack/boilerplate.go.txt
chmod -x "$CODEGEN_DIR/generate-groups.sh"

rsync -av "output_base/github.com/pivotal/monitoring-indicator-protocol/k8s/pkg/apis/" "$REPO_ROOT/k8s/pkg/apis/"
rsync -av --delete "output_base/github.com/pivotal/monitoring-indicator-protocol/k8s/pkg/client/" "$REPO_ROOT/k8s/pkg/client/"
rm -rf output_base

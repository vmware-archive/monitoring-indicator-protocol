#!/usr/bin/env bash

set -e

pushd ~/workspace > /dev/null
  rm -rf monitoring-indicator-protocol/bosh-release/src/github.com/pivotal/
  # update code in bosh-release src directory to have the latest locally
  mkdir -p monitoring-indicator-protocol/bosh-release/src/github.com/pivotal/
  rsync -Rr ./monitoring-indicator-protocol/ ./monitoring-indicator-protocol/bosh-release/src/github.com/pivotal/

  BUILD_NUMBER=test-$(date +"%s")

  pushd monitoring-indicator-protocol/bosh-release > /dev/null
    bosh -n create-release --sha2 --force --version "${BUILD_NUMBER}" --tarball ./"indicator-release-${BUILD_NUMBER}".tgz
  popd > /dev/null
popd > /dev/null

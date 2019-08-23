#!/usr/bin/env bash

DEPLOYMENT_NAME=cf-01234567890123456789
PROMETHEUS_URI=https://metric-store.madlamp.cf-denver.com
UAA_URI=https://uaa.madlamp.cf-denver.com
UAA_CLIENT_ID=healthwatch_admin

set -e

pushd ~/workspace > /dev/null
    mkdir -p monitoring-indicator-protocol/bosh-release/src/github.com/pivotal/
    rsync -Rr ./monitoring-indicator-protocol/ ./monitoring-indicator-protocol/bosh-release/src/github.com/pivotal/

    BUILD_NUMBER=test-$(date +"%s")

    pushd monitoring-indicator-protocol/bosh-release > /dev/null

      cat << EOF > config/private.yml
---
blobstore:
  options:
    credentials_source: static
    json_key: |
      ${SERVICE_ACCOUNT}
EOF

      bosh -n create-release --sha2 --force --version ${BUILD_NUMBER}
      bosh -n upload-release --fix

      bosh update-runtime-config -n \
        --name indicator-document-registration-agent \
        --var=indicator-protocol-version=${BUILD_NUMBER} \
        manifests/agent_runtime_config.yml

    UAA_CLIENT_SECRET=$(credhub g -n /bosh-madlamp/cf-01234567890123456789/uaa_clients_cc-service-dashboards_secret -j | jq -r .value)

    echo -e "${EXAMPLE_REPO_PRIVATE_KEY}" > ./sample-repo-key

    bosh -n -d indicator-protocol deploy \
      manifests/manifest.yml \
      -o ops_files/add-examples-git-source.yml \
      -o ops_files/configure-status-controller.yml \
      -o ops_files/set-indicator-protocol-version.yml \
      -v prometheus_uri=$PROMETHEUS_URI \
      -v uaa_uri=$UAA_URI \
      -v uaa_client_id=$UAA_CLIENT_ID \
      -v uaa_client_secret=$UAA_CLIENT_SECRET \
      -v indicator-protocol-version=$BUILD_NUMBER \
      --var-file=patch_repo_private_key=./sample-repo-key

    rm ./sample-repo-key

    popd > /dev/null
popd > /dev/null

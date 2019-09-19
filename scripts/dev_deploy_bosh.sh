#!/usr/bin/env bash

DEPLOYMENT_NAME=cf-01234567890123456789
PROMETHEUS_URI=https://metric-store.madlamp.cf-denver.com
UAA_URI=https://uaa.madlamp.cf-denver.com
UAA_CLIENT_ID=healthwatch_admin

set -e
#source ~/workspace/denver-bash-it/custom/environment-targeting.bash
#target madlamp

pushd ~/workspace > /dev/null
  # update code in bosh-release src directory to have the latest locally
  mkdir -p monitoring-indicator-protocol/bosh-release/src/github.com/pivotal/
  rsync -Rr --exclude ./monitoring-indicator-protocol/bosh-release \
        ./monitoring-indicator-protocol/ ./monitoring-indicator-protocol/bosh-release/src/github.com/pivotal/

  BUILD_NUMBER=test-$(date +"%s")

  pushd monitoring-indicator-protocol/bosh-release > /dev/null
    # prep for bosh deploy
    bosh -n create-release --sha2 --force --version "${BUILD_NUMBER}"
    bosh -n upload-release --fix

    bosh update-runtime-config -n \
      --name indicator-document-registration-agent \
      --var=indicator-protocol-version="${BUILD_NUMBER}" \
      manifests/agent_runtime_config.yml

    UAA_CLIENT_SECRET=$(credhub g -n /bosh-madlamp/cf-01234567890123456789/uaa_clients_cc-service-dashboards_secret -j | jq -r .value)

    # bosh deploy. ops_file params are optional
    bosh -n -d indicator-protocol deploy \
      manifests/manifest.yml \
      -o ops_files/add-examples-git-source.yml \
      -o ops_files/configure-status-controller.yml \
      -o ops_files/set-indicator-protocol-version.yml \
      -v prometheus_uri=$PROMETHEUS_URI \
      -v uaa_uri=$UAA_URI \
      -v uaa_client_id=$UAA_CLIENT_ID \
      -v uaa_client_secret="$UAA_CLIENT_SECRET" \
      -v indicator-protocol-version="$BUILD_NUMBER" \
      -v system_domain="sys.madlamp.cf-denver.com"

    # uncomment the following to re-deploy PAS if you have changes to the agent
    # CF_DEPLOYMENT_NAME=$(bosh deployments --json | jq .Tables[0].Rows | jq '.[] | select( .name | contains("cf"))' | jq .name -r)
    # bosh -n -d $CF_DEPLOYMENT_NAME manifest > temp.yml
    # bosh -n -d $CF_DEPLOYMENT_NAME deploy temp.yml

  popd > /dev/null
popd > /dev/null

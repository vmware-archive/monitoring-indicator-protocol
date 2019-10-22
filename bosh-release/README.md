## Monitoring Indicator Protocol Bosh Release

This directory contains the files needed to deploy indicator protocol components
using Bosh.

### Deploying

The [releases](https://github.com/pivotal/monitoring-indicator-protocol/releases)
page contains a Bosh release for each version of Indicator Protocol.
To get started, you can download the latest one.
Once downloaded, you can upload it to Bosh like so
```bash
bosh upload-release indicator-protocol-bosh-0.7.10.tgz
```

You will need to add a runtime config for the registration agent:
```bash
bosh update-runtime-config -n \
  --name indicator-document-registration-agent \
  --var=indicator-protocol-version=$(bosh releases | grep indicator-protocol -m1 | cut -f2) \
  monitoring-indicator-protocol/bosh-release/manifests/agent_runtime_config.yml
```

Simple deploy (without status updates or external patch/document Git sources):
```bash
bosh -n -d indicator-registry deploy \
    monitoring-indicator-protocol/bosh-release/manifests/manifest.yml
```

#### Configuring

Additionally, there are two ops files included in the repository:

1. `add-examples-git-source.yml` configures the registry to look for indicator
documents and patches in
[this](https://github.com/pivotal/indicator-protocol-examples)
GitHub repository.
1. `configure-status-controller` configures the indicator status controller to
communicate with a prometheus compliant datastore using UAA client credentials.

Complete deploy:
```bash
bosh -n -d indicator-protocol deploy \
    monitoring-indicator-protocol/bosh-release/manifests/manifest.yml \
    -o monitoring-indicator-protocol/bosh-release/ops_files/add-examples-git-source.yml \
    -o monitoring-indicator-protocol/bosh-release/ops_files/configure-status-controller.yml \
    -v prometheus_uri=$PROMETHEUS_URI \
    -v uaa_uri=$UAA_URI \
    -v uaa_client_id=$UAA_CLIENT_ID \
    -v uaa_client_secret=$UAA_CLIENT_SECRET
```

#### Updating dependencies
We use package vendoring for the Golang package (https://bosh.io/docs/package-vendoring/#vendor). 
We have a file called “bosh-release/packages/golang-1.*-linux/spec.lock” that has the fingerprint of the golang version we are using.
https://github.com/bosh-packages/golang-release has the golang release.

* Find the most recent release in https://github.com/bosh-packages/golang-release/tree/master/releases/golang
* Use the fingerprint for the linux release of the go minor version you want
* Once you’ve changed the spec.lock file, you can deploy it to some acceptance environment, and then check the package fingerprint, as well as looking at the release locally and looking at the go patch version there
* Follow the instruction in the golang-release readme about vendoring the package
    * `cd ~/workspace/golang-release && git pull`
    * `cd ~/workspace/monitoring-indicator-protocol/bosh-release`
    * run `bosh vendor-package golang-1.13-linux ~/workspace/golang-release`
    * run `../scripts/dev_deploy_bosh.sh` to test that it worked


The tile will pull in the latest bosh release, so there is no need to do any additional work


#!/usr/bin/env bash
# This script runs all the tests, after ensuring that everything is set up
# at least roughly the way it will be set up in CI. It takes a while to run,
# so it is meant to be run before commiting to increase the chance of CI passing.


SCRIPT=`realpath $0`
SCRIPTDIR=`dirname $SCRIPT`


# So that we have access to `target` which is not on the path.
. ~/.bash_profile

pushd $SCRIPTDIR/.. > /dev/null
    ./scripts/test.sh local

    target madlamp
    gcloud container clusters get-credentials mip-development -z us-central1-a
    kubectl config use-context gke_cf-indicators-team_us-central1-a_mip-development

    ./hack/update-codegen.sh
    ./scripts/dev_deploy_k8s.sh

    ./scripts/test.sh e2e
popd > /dev/null

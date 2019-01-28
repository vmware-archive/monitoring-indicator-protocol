#!/bin/bash
set -ef

set -eu

lpass ls > /dev/null # check that we're logged in

fly -t superpipe set-pipeline -p indicator-protocol \
    -c "pipelines/indicator_protocol.yml" \
    --load-vars-from <(lpass show --notes "Shared-Event Producer (Pivotal Only)/concourse.yml")

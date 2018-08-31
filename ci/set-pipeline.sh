#!/bin/bash
set -ef

if [[ $# -ne 1 ]];
then
    echo "Incorrect usage: $0 pipeline-name"
    exit 1
fi

set -eu

lpass ls > /dev/null # check that we're logged in

echo setting pipeline for "$1"
fly -t superpipe set-pipeline -p "$1" \
    -c "pipelines/$1.yml" \
    --load-vars-from <(lpass show --notes "Shared-Event Producer (Pivotal Only)/concourse.yml")

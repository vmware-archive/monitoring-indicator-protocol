#!/bin/bash
set -ef

set -eu

lpass ls > /dev/null # check that we're logged in

fly -t indipro set-pipeline -p indicator-protocol \
    -c "pipelines/indicator_protocol.yml"

#!/bin/bash

set -efu

fly -t indipro set-pipeline -p indicator-protocol-v0.7 \
    -c "ci/pipelines/indicator_protocol.yml"

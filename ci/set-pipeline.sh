#!/bin/bash

set -efu

fly -t indipro set-pipeline -p indicator-protocol \
    -c "pipelines/indicator_protocol.yml"

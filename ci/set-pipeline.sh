#!/bin/bash

set -efu

fly --target indipro set-pipeline \
    --pipeline indicator-protocol-v0.7 \
    --config "ci/pipeline.yml"

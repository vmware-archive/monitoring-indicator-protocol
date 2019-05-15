#!/bin/bash

set -efu

fly --target indipro set-pipeline \
    --pipeline indicator-protocol \
    --config "pipelines/indicator_protocol.yml"

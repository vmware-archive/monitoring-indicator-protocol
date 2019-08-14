#!/bin/bash

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )/.." > /dev/null 2>&1 && pwd )"

target_dir="$DIR/bosh-release/src/github.com/pivotal/monitoring-indicator-protocol/"
mkdir -p "$target_dir"
rsync -a "$DIR/" "$target_dir"

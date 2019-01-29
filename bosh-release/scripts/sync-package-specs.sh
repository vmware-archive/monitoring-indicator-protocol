#!/bin/bash
set -e

go get github.com/vito/gosub

if [ ! $(which gosub) ]; then
    echo "Gosub required to update dependencies in bosh/*/spec files."
    echo 'Please install with `go get github.com/vito/gosub`'
    exit 1
fi

function sync_package() {
  bosh_pkg=${1}

  shift

  (
    set -e

    cd packages/${bosh_pkg}

    {
      cat spec | grep -v '# gosub'
      gosub list "$@" | \
        sed -e 's|\(.*\)|- \1/*.go # gosub|g'
    } > spec.new

    mv spec.new spec

  )
}

sync_package indicator-registry -app github.com/pivotal/indicator-protocol/cmd/registry &
sync_package indicator-registration-agent -app github.com/pivotal/indicator-protocol/cmd/registry_agent &

wait

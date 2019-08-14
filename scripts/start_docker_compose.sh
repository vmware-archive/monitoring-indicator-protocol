#!/usr/bin/env bash

pushd docker-compose
  ./generate_certs.sh
  docker-compose up

popd

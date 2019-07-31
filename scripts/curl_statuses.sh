#!/usr/bin/env bash

curl -s https://localhost:8091/v1/indicator-documents -k --cert test_fixtures/client.pem --key test_fixtures/client.key | jq '.[] | .indicators | .[] | {name: .name, status: .status}'

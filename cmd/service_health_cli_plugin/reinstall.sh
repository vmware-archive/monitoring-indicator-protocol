#/usr/bin/env bash

go build
cf install-plugin service_health_cli_plugin -f
cf install-plugin service_health_cli_plugin -f
cf service-health my-rabbit

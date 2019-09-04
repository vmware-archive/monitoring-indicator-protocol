#!/usr/bin/env bash

if [ ! -f ~/Documents/sangabriel.json ]; then
  echo "Missing ~/Documents/sangabriel.json"
  exit 1
fi

jq -r .ops_manager_private_key ~/Documents/sangabriel.json  > /tmp/ops_manager_private_key.pem


export OM_USERNAME=$(jq .ops_manager.username -r ~/Documents/sangabriel.json )
export OM_PASSWORD=$(jq .ops_manager.password -r ~/Documents/sangabriel.json )
export OM_TARGET=$(jq .ops_manager.url -r  ~/Documents/sangabriel.json )

CF_PASSWORD=$(om -k credentials -p cf -c .uaa.admin_credentials -f password )

cf login -a api.sys.sangabriel.cf-app.com \
    -u admin -p "${CF_PASSWORD}" \
    -o system -s system \
    --skip-ssl-validation

''

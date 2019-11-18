#!/bin/bash

set -euxo pipefail
export VAULT_ADDR='http://127.0.0.1:8200'
export PLUGIN_NAME='pachyderm'
export PLUGIN_PATH=$PLUGIN_NAME

# Build it
pachctl version
which aws || pip install awscli --upgrade --user
if [[ "$(pachctl enterprise get-state)" = "No Pachyderm Enterprise token was found" ]]; then
  pachctl enterprise activate  $(aws s3 cp s3://pachyderm-engineering/test_enterprise_activation_code.txt -)
fi
if ! pachctl auth list-admins; then
  yes | pachctl auth activate -u admin
fi

echo 'root' | vault login -

# Clean up from last run
vault secrets disable $PLUGIN_PATH

# Enable the plugin
export SHASUM=$(shasum -a 256 "/tmp/vault-plugins/$PLUGIN_NAME" | cut -d " " -f1)
echo $SHASUM
vault write sys/plugins/catalog/$PLUGIN_NAME sha_256="$SHASUM" command="$PLUGIN_NAME"
vault secrets enable -path=$PLUGIN_PATH -plugin-name=$PLUGIN_NAME plugin

# Set the admin token vault will use to create user creds
yes | pachctl auth login -u admin
export ADMIN_TOKEN=$(cat ~/.pachyderm/config.json | jq -r .v1.session_token)
echo $ADMIN_TOKEN
vault write $PLUGIN_PATH/config \
    admin_token="${ADMIN_TOKEN}" \
	pachd_address="${PACHD_ADDRESS:-127.0.0.1}:30650"
go build -mod=vendor -o /tmp/vault-plugins/$PLUGIN_NAME src/plugin/vault/main.go

# Test login (failure/success):
vault write -f $PLUGIN_PATH/login/github:jdoliner
vault write -f $PLUGIN_PATH/login/github:JoeyZwicker ttl=125s

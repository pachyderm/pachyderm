#!/bin/bash

set -euxo pipefail
export VAULT_ADDR='http://127.0.0.1:8200'
export PLUGIN_NAME='pachyderm'
export PLUGIN_PATH=$PLUGIN_NAME

# Build it

pachctl version
# Remove the admin I'm going to add to make sure my test has a side effect
pachctl auth modify-admins --remove daffyduck || true

go build -o /tmp/vault-plugins/$PLUGIN_NAME src/plugin/vault/main.go 

echo 'root' | vault login -

# Clean up from last run
vault auth disable $PLUGIN_PATH 

# Enable the plugin
export SHASUM=$(shasum -a 256 "/tmp/vault-plugins/$PLUGIN_NAME" | cut -d " " -f1)
echo $SHASUM
vault write sys/plugins/catalog/$PLUGIN_NAME sha_256="$SHASUM" command="$PLUGIN_NAME"
vault auth enable -path=$PLUGIN_PATH -plugin-name=$PLUGIN_NAME plugin

# Test login before admin token is set
vault write auth/$PLUGIN_PATH/login username=tweetybird || true

# Set the admin token vault will use to create user creds
export ADMIN_TOKEN=$(cat ~/.pachyderm/config.json | jq -r .v1.session_token)
echo $ADMIN_TOKEN
vault write auth/$PLUGIN_PATH/config \
    admin_token="${ADMIN_TOKEN}"

# Test login (failure/success):
vault write auth/$PLUGIN_PATH/login username=bogusgithubusername || true
vault write auth/$PLUGIN_PATH/login username=daffyduck

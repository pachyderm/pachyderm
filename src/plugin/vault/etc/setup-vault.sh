#!/bin/bash

set -euxo pipefail

# Make sure Pachyderm is running
which pachctl
pachctl version

export VAULT_ADDR='http://127.0.0.1:8200'
export PLUGIN_NAME='pachyderm'

# Make sure vault binary is present
which vault

echo "going to login to vault"
echo 'root' | vault login -
echo "logged into vault"

# Remove the old plugin binary
set +o pipefail
rm /tmp/vault-plugins/$PLUGIN_NAME || true
set -o pipefail

# Build the plugin binary
SCRIPT_DIR=$(dirname "${0}")
go build -o /tmp/vault-plugins/$PLUGIN_NAME ${SCRIPT_DIR}/..

# Disable the existing plugin (i.e. kill the plugin process)
vault secrets disable $PLUGIN_NAME

# Re-enable the plugin (i.e. start the new plugin process)
export SHASUM=$(shasum -a 256 "/tmp/vault-plugins/$PLUGIN_NAME" | cut -d " " -f1)
echo $SHASUM
vault write sys/plugins/catalog/$PLUGIN_NAME sha_256="$SHASUM" command="$PLUGIN_NAME"
vault secrets enable -path=$PLUGIN_NAME -plugin-name=$PLUGIN_NAME plugin

# Make sure Pachyderm enterprise and auth are enabled
which aws || pip install awscli --upgrade --user
if [[ "$(pachctl enterprise get-state)" = "No Pachyderm Enterprise token was found" ]]; then
  # Don't print token to stdout
  # This is very important, or we'd leak it in our CI logs
  set +x
  pachctl enterprise activate  $(aws s3 cp s3://pachyderm-engineering/test_enterprise_activation_code.txt -)
  set -x
fi

# Connect pachyderm plugin to cluster
if ! pachctl auth list-admins; then
  admin="admin"
  echo "${admin}" | pachctl auth activate
else
  admin="$( pachctl auth list-admins | grep github: | head -n 1)"
  admin="${admin#github:}"
  echo "${admin}" | pachctl auth login
fi

vault write pachyderm/config admin_token=$(pachctl auth get-auth-token "github:${admin}" | grep Token | awk '{print $2}') pachd_address=$(minikube ip):30650

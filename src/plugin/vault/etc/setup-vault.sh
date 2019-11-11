#!/bin/bash

set -euxo pipefail

# Make sure Pachyderm is running and auth is activated
which pachctl
pachctl version
pachctl auth whoami

# Make sure vault binary is present
which vault

# generate an auth token for one of Pachyderm's existing admins
admin="$(
  pachctl auth list-admins \
    | grep 'github:' | head -n 1 | sed 's/^github://'
)"
ADMIN_TOKEN="$(pachctl auth get-auth-token "github:${admin}" | grep Token | awk '{print $2}')"

# Login to vault, so we can disable the previous Pachyderm plugin (if any is running)
export VAULT_ADDR='http://127.0.0.1:8200'
export PLUGIN_NAME='pachyderm'

echo "going to login to vault"
echo 'root' | vault login -
echo "logged into vault"

# If the Pachyderm plugin is running, disable it. Note that we re-grant the
# plugin access to Pachyderm so that it can revoke any existing tokens
if vault read pachyderm/version/client-only; then
  vault write pachyderm/config \
        admin_token="${ADMIN_TOKEN}" \
        pachd_address="${PACHD_ADDRESS:-127.0.0.1:30650}" \
        ttl=5m # optional
  vault secrets disable $PLUGIN_NAME
fi

# Remove the old plugin binary
set +o pipefail
rm /tmp/vault-plugins/$PLUGIN_NAME || true
set -o pipefail

# Build the plugin binary
(
  cd ${GOPATH}/src/github.com/pachyderm/pachyderm/src/plugin/vault && \
  make plugin
)
mkdir -p /tmp/vault-plugins
cp ${GOPATH}/bin/pachyderm-plugin /tmp/vault-plugins/${PLUGIN_NAME}

# Re-enable the plugin (i.e. start the new plugin process)
export SHASUM=$(shasum -a 256 "/tmp/vault-plugins/$PLUGIN_NAME" | cut -d " " -f1)
echo $SHASUM
vault write sys/plugins/catalog/$PLUGIN_NAME sha_256="$SHASUM" command="$PLUGIN_NAME"
vault secrets enable -path=$PLUGIN_NAME -plugin-name=$PLUGIN_NAME plugin

vault write pachyderm/config admin_token=${ADMIN_TOKEN} pachd_address=$(minikube ip):30650

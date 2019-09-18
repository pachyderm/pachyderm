#!/bin/bash

set -euxo pipefail

# Make sure Pachyderm is running
which pachctl
pachctl version

# Make sure vault binary is present
which vault

# Make sure Pachyderm enterprise and auth are enabled
which aws || pip install awscli --upgrade --user
if [[ "$(pachctl enterprise get-state)" = "No Pachyderm Enterprise token was found" ]]; then
  # Don't print token to stdout
  # This is very important, or we'd leak it in our CI logs
  set +x
  pachctl enterprise activate $(aws s3 cp s3://pachyderm-engineering/test_enterprise_activation_code.txt -)
  set -x
fi

# Activate Pachyderm auth, if needed, and get a Pachyderm admin token
if ! pachctl auth list-admins; then
  admin="admin"
  echo "${admin}" | pachctl auth activate
elif pachctl auth list-admins | grep "github:"; then
  admin="$( pachctl auth list-admins | grep 'github:' | head -n 1)"
  admin="${admin#github:}"
  echo "${admin}" | pachctl auth login
else
  echo "Could not find a github user to log in as. Cannot get admin token"
  exit 1
fi
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

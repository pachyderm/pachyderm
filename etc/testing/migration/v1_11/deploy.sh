#!/bin/bash
# deploy.sh deploys a pachyderm 1.11.9 cluster (the first release with auth extract/restore)

SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"
source "${SCRIPT_DIR}/../../../govars.sh"

set -x


# Install old version of pachctl, for migration tests
if [[ ! -f "${GOBIN}/pachctl_1_11" ]]; then
  curl -Ls https://github.com/pachyderm/pachyderm/releases/download/v1.11.9/pachctl_1.11.9_linux_amd64.tar.gz \
    | tar -xz pachctl_1.11.9_linux_amd64/pachctl -O >"${GOBIN}/pachctl_1_11"
  chmod +x "${GOBIN}/pachctl_1_11"
fi

# (If 1.11 is already deployed, we're done. Otherwise, undeploy + deploy)
if ! grep . <( kubectl get po -l suite=pachyderm 2>/dev/null ) \
  || [[ "$(pachctl version | grep pachd | awk '{print $2}' )" != "1.11.9" ]]; then
  # Clear any existing pachyderm cluster
  if pachctl version --timeout=2s &>/dev/null; then
    ( (yes | pachctl delete all) && pachctl undeploy) || {
      echo "Error: could not clear existing Pachyderm cluster. Giving up..."
      exit 1
    }
  fi

  # Deploy Pachyderm 1.11 cluster
  pachctl_1_11 deploy local

  # Wait for pachyderm to come up
  HERE="$(dirname "${0}")"
  until timeout 1s "${HERE}/../../../kube/check_ready.sh" app=pachd; do sleep 1; done
fi

#!/bin/bash
# deploy.sh deploys a pachyderm 1.11.9 cluster (the first release with auth extract/restore)

set -x

# Install old version of pachctl, for migration tests
if [[ ! -f "${GOPATH}/bin/pachctl_1_11" ]]; then
  curl -Ls https://github.com/pachyderm/pachyderm/releases/download/v1.11.9/pachctl_1.11.9_linux_amd64.tar.gz \
    | tar -xz pachctl_1.11.9_linux_amd64/pachctl -O >"${GOPATH}/bin/pachctl_1_11"
  chmod +x "${GOPATH}/bin/pachctl_1_11"
fi

# (If 1.11 is already deployed, we're done. Otherwise, undeploy + deploy)
if ! grep . <( kubectl get po -l suite=pachyderm 2>/dev/null ) \
  || [[ "$(pachctl version | grep pachd | awk '{print $2}' )" != "1.11.9" ]]; then
  # Clear any existing pachyderm cluster
  if pachctl version --timeout=2s &>/dev/null; then
    ( (yes | pachctl delete-all) && pachctl undeploy) || {
      echo "Error: could not clear existing Pachyderm cluster. Giving up..."
      exit 1
    }
  fi

  # Deploy Pachyderm 1.7 cluster
  pachctl_1_11 deploy local

  # Wait for pachyderm to come up
  set +x
  HERE="$(dirname "${0}")"
  PACHCTL=pachctl_1_11 "${HERE}/../../../kube/wait_for_startup.sh"
  set -x
fi

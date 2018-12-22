#!/bin/bash
# deploy.sh deploys a pachyderm 1.7 cluster

set -x

# Install old version of pachctl, for migration tests
if [[ ! -f ${GOPATH}/bin/pachctl_1_7 ]]; then
  curl -Ls https://github.com/pachyderm/pachyderm/releases/download/v1.7.11/pachctl_1.7.11_linux_amd64.tar.gz \
    | tar -xz pachctl_1.7.11_linux_amd64/pachctl -O >${GOPATH}/bin/pachctl_1_7
  chmod +x ${GOPATH}/bin/pachctl_1_7
fi

# (If 1.7 is already deployed, we're done. Otherwise, undeploy + deploy)
if ! grep . <( kubectl get po -l suite=pachyderm 2>/dev/null ) \
  || [[ "$(pachctl version | grep pachd | awk '{print $2}' )" != "1.7.11" ]]; then
  # Clear any existing pachyderm cluster
  if pachctl version --timeout=2s 2>&1 >/dev/null; then
    yes | pachctl delete-all && pachctl undeploy || {
      echo "Error: could not clear existing Pachyderm cluster. Giving up..."
      exit 1
    }
  fi

  # Deploy Pachyderm 1.7 cluster
  pachctl_1_7 deploy local

  # Wait for pachyderm to come up
  set +x
  HERE="$(dirname "${0}")"
  PACHCTL=pachctl_1_7 "${HERE}/../../../kube/wait_for_startup.sh"
  set -x
fi

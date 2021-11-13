#!/bin/bash

set -euxo pipefail

mkdir -p "${HOME}/go/bin"
export PATH=$PATH:/usr/local/go/bin:$HOME/go/bin
export GOPATH="${HOME}/go"

# Install go.
sudo rm -rf /usr/local/go
curl -L https://golang.org/dl/go1.17.3.linux-amd64.tar.gz | sudo tar xzf - -C /usr/local/
go version

# install hubcli
pushd "${HOME}/go/bin"
rm -f hubcli
wget https://github.com/pachyderm/hubcli/releases/download/0.0.2/hubcli
chmod a+x hubcli
popd

# Install goreleaser.
GORELEASER_VERSION=0.169.0
curl -L "https://github.com/goreleaser/goreleaser/releases/download/v${GORELEASER_VERSION}/goreleaser_Linux_x86_64.tar.gz" \
    | tar xzf - -C "${HOME}/go/bin" goreleaser

# Build pachctl.
make install
pachctl version --client-only

# Set version for docker builds.
VERSION="$(pachctl version --client-only)"
export VERSION

# Build and push docker images.
make docker-build
make docker-push

# Create a workspace running the image we just built.
for _ in {1..3}; do
  hubcli \
    --endpoint https://hub.pachyderm.com/api/graphql \
    --apikey "${HUB_API_KEY}" \
    --op create-workspace-and-wait \
    --orgid 2193 \
    --loglevel trace \
    --infofile workspace.json \
    --version "${VERSION}" \
    --expiration 2h \
    --description "${CIRCLE_BUILD_URL}" \
    --prefix "ci-" \
    && break
done

# Print client and server versions, for debugging.
pachctl version

# Run load tests.
set +e
pachctl run pfs-load-test "${@}"
if [ "${?}" -ne 0 ]; then
	pachctl debug dump /tmp/debug-dump
	exit 1
fi
pachctl run pps-load-test "${@}"
if [ "${?}" -ne 0 ]; then
	pachctl debug dump /tmp/debug-dump
	exit 1
fi
set -e
pachctl debug dump /tmp/debug-dump

# Delete the workspace.  We don't do this in a "trap ... exit" statement so that you can log into
# the workspace and debug it if the load tests fail.  Hub will automatically clean up the workspace
# at its expiration time set above.
hubcli \
  --endpoint https://hub.pachyderm.com/api/graphql \
  --apikey "${HUB_API_KEY}" \
  --op delete-workspace \
  --loglevel trace \
  --infofile workspace.json

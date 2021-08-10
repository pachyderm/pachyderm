#!/bin/bash

set -euxo pipefail

mkdir -p $HOME/go/bin
export PATH=$PATH:/usr/local/go/bin:$HOME/go/bin

# Install go.
sudo rm -rf /usr/local/go
curl -L https://golang.org/dl/go1.16.6.linux-amd64.tar.gz | sudo tar xzf - -C /usr/local/
go version

# install hubcli
pushd $HOME/go/bin
rm -f hubcli
wget https://github.com/pachyderm/hubcli/releases/download/0.0.2/hubcli
chmod a+x hubcli
popd

# Install goreleaser.
GORELEASER_VERSION=0.169.0
curl -L https://github.com/goreleaser/goreleaser/releases/download/v${GORELEASER_VERSION}/goreleaser_Linux_x86_64.tar.gz \
    | tar xzf - -C $HOME/go/bin goreleaser

# Build pachctl.
make install

# set up version for docker builds
export VERSION=$(pachctl version --client-only)

# build docker images
# goreleaser release -p 1 --snapshot --skip-publish --rm-dist -f goreleaser/docker.yml
# docker tag pachyderm/pachd pachyderm/pachd:$VERSION
# docker tag pachyderm/worker pachyderm/worker:$VERSION
# docker tag pachyderm/pachctl pachyderm/pachctl:$VERSION
# docker push pachyderm/pachd:$VERSION
# docker push pachyderm/worker:$VERSION
# docker push pachyderm/pachctl:$VERSION
make docker-build
make docker-push

# Create a workspace running the image we just built.
hubcli --endpoint https://hub.pachyderm.com/api/graphql --apikey $HUB_API_KEY --op create-workspace-and-wait --orgid 2193 --loglevel trace --infofile workspace.json --version $VERSION --expiration 2h --description $CIRCLE_BUILD_URL --prefix "ci-"

# Print client and server versions, for debugging.
pachctl version

# Run load tests.
pachctl run pfs-load-test
pachctl run pps-load-test

# Delete the workspace.  We don't do this in a "trap ... exit" statement so that you can log into
# the workspace and debug it if the load tests fail.  Hub will automatically clean up the workspace
# at its expiration time set above.
hubcli --endpoint https://hub.pachyderm.com/api/graphql --apikey $HUB_API_KEY --op delete-workspace --loglevel trace --infofile workspace.json

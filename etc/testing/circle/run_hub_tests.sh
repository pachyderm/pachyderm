#!/bin/bash

set -euxo pipefail

mkdir -p $HOME/go/bin
export PATH=$PATH:$HOME/go/bin

# install hubcli
pushd $HOME/go/bin
rm -f hubcli
wget https://github.com/pachyderm/hubcli/releases/download/0.0.2/hubcli
chmod a+x hubcli
popd

# install goreleaser
GORELEASER_VERSION=0.169.0
curl -L https://github.com/goreleaser/goreleaser/releases/download/v${GORELEASER_VERSION}/goreleaser_Linux_x86_64.tar.gz \
    | tar xzf - -C $HOME/go/bin goreleaser

# TODO: upgrade go here (as per install.sh)

# setup environment for build
export VERSION_ADDITIONAL="-$(git log --pretty=format:%H | head -n 1)"
export CLIENT_ADDITIONAL_VERSION="github.com/pachyderm/pachyderm/v2/src/version.AdditionalVersion=${VERSION_ADDITIONAL}"
export GC_FLAGS="all=-trimpath=${PWD}"
export LD_FLAGS="-X ${CLIENT_ADDITIONAL_VERSION}"

# build pachctl
make install

# set up version for docker builds
export VERSION=$(pachctl version --client-only)

# build docker images
goreleaser release -p 1 --snapshot --skip-publish --rm-dist -f goreleaser/docker.yml
docker tag pachyderm/pachd pachyderm/pachd:$VERSION
docker tag pachyderm/worker pachyderm/worker:$VERSION
docker tag pachyderm/pachctl pachyderm/pachctl:$VERSION
docker push pachyderm/pachd:$VERSION
docker push pachyderm/worker:$VERSION
docker push pachyderm/pachctl:$VERSION

# create a workspace
hubcli --endpoint https://hub.pachyderm.com/api/graphql --apikey $HUB_API_KEY --op create-workspace-and-wait --orgid 2193 --loglevel trace --infofile workspace.json --version $VERSION --expiration 2h --description $CIRCLE_BUILD_URL --prefix "ci-"

# print versions for debugging
pachctl version

# run tests against hub
pachctl run pfs-load-test
pachctl run pps-load-test

# delete the workspace
hubcli --endpoint https://hub.pachyderm.com/api/graphql --apikey $HUB_API_KEY --op delete-workspace --loglevel trace --infofile workspace.json

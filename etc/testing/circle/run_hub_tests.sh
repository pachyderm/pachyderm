#!/bin/bash

set -euxo pipefail

# install hubcli
wget https://github.com/pachyderm/hubcli/releases/download/v0.0.1-beta.1/hubcli
chmod a+x hubcli

# install pachctl and friends
make install
VERSION=$(pachctl version --client-only)
git config user.email "donotreply@pachyderm.com"
git config user.name "anonymous"
git tag -f -am "Circle CI test v$VERSION" v"$VERSION"

# push a docker image
make docker-build

# create a workspace
./hubcli --endpoint https://hub.pachyderm.com/api/graphql --apikey $HUB_API_KEY --op create-workspace-and-wait --orgid 2193 --loglevel trace --infofile workspace.json --version $VERSION

# run tests against hub
pachctl run pfs-load-test
pachctl run pps-load-test

# delete the workspace
./hubcli --endpoint https://hub.pachyderm.com/api/graphql --apikey $HUB_API_KEY --op delete-workspace --loglevel trace --infofile workspace.json

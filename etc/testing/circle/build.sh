#!/bin/bash

set -ex

# shellcheck disable=SC1090
source "$(dirname "$0")/env.sh"

## pachctl build
go version
make install
VERSION=$(pachctl version --client-only)

## local docker build


# then actually do the build
git config user.email "donotreply@pachyderm.com"
git config user.name "anonymous"
git tag -f -am "Circle CI test v$VERSION" v"$VERSION"
make docker-build

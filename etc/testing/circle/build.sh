#!/bin/bash

set -ex

source "$(dirname "$0")/env.sh"

make install
VERSION=$(pachctl version --client-only)
git config user.email "donotreply@pachyderm.com"
git config user.name "anonymous"
git tag -f -am "Circle CI test v$VERSION" v"$VERSION"
make docker-build

kind load docker-image pachyderm/pachd:local
kind load docker-image pachyderm/pachctl:local
kind load docker-image pachyderm/worker:local

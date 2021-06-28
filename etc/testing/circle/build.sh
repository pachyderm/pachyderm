#!/bin/bash

set -ex

source "$(dirname "$0")/env.sh"

make install
VERSION=$(pachctl version --client-only)
git config user.email "donotreply@pachyderm.com"
git config user.name "anonymous"
git tag -f -am "Circle CI test v$VERSION" v"$VERSION"
make docker-build

mkdir images
docker export -o images/pachd.tar pachd
docker export -o images/pachctl.tar pachctl
docker export -o images/worker.tar worker
docker export -o images/worker_init.tar worker_init


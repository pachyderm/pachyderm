#!/bin/bash

set -ex

make docker-build
mkdir -p ~/cached-deps
docker save pachyderm/pachd:local | gzip > ~/cached-deps/pachd.tar.gz
docker save pachyderm/worker:local | gzip > ~/cached-deps/worker.tar.gz

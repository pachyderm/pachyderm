#!/bin/bash

set -ex

make install docker-build
docker push pachyderm/pachd:local pachyderm/pachd:$(pachctl version --client-only)
docker push pachyderm/pachd:local pachyderm/worker:$(pachctl version --client-only)

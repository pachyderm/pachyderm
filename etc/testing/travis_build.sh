#!/bin/bash

set -ex

docker login -u pachydermbuildbot -p ${DOCKER_PWD}

make install docker-build
version=$(pachctl version --client-only)
docker tag pachyderm/pachd:local pachyderm/pachd:${version}
docker push pachyderm/pachd:${version}
docker tag pachyderm/worker:local pachyderm/worker:${version}
docker push pachyderm/worker:${version}

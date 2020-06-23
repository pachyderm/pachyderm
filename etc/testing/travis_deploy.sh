#!/bin/bash

set -ex

# We can't run the build step if there's no access to the secret env vars
if [[ "$TRAVIS_SECURE_ENV_VARS" == "true" ]]; then
    docker load < ~/cached-deps/pachd.tar.gz
    docker load < ~/cached-deps/worker.tar.gz

    make install
    docker login -u pachydermbuildbot -p "${DOCKER_PWD}"
    version=$(pachctl version --client-only)
    docker tag "pachyderm/pachd:local" "pachyderm/pachd:${version}"
    docker push "pachyderm/pachd:${version}"
    docker tag "pachyderm/worker:local" "pachyderm/worker:${version}"
    docker push "pachyderm/worker:${version}"
fi

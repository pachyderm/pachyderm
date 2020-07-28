#!/bin/bash

set -ex

# We can't run the build step if there's no access to the secret env vars
if [[ $TRAVIS_SECURE_ENV_VARS == "true" ]]; then
    docker login -u pachydermbuildbot -p "${DOCKER_PWD}"
    make install docker-build
    version=$(pachctl version --client-only)
    docker tag "pachyderm/pachd:local" "pachyderm/pachd:${version}"
    docker push "pachyderm/pachd:${version}"
    docker tag "pachyderm/worker:local" "pachyderm/worker:${version}"
    docker push "pachyderm/worker:${version}"

    # Push pipeline build images
    make docker-push-pipeline-build
fi

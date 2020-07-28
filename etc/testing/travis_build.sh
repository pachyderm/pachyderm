#!/bin/bash

set -ex

# Lint as basically the first thing we do, this catches a lot of errors that
# are otherwise buried deep in the Travis logs and gives developers faster
# feedback than waiting ~20 minutes and then digging through the logs.
make lint

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

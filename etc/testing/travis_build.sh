#!/bin/bash

set -ex

# We can't run the build step if there's no access to the secret env vars
if [[ "$TRAVIS_SECURE_ENV_VARS" == "true" ]]; then
    docker login -u pachydermbuildbot -p "${DOCKER_PWD}"

    # Load saved docker image layers, to avoid having to repeatedly apt-get
    # stuff every time we build
    docker load -i ${HOME}/docker_images/images.tar || true

    make install docker-build
    version=$(pachctl version --client-only)
    docker tag "pachyderm/pachd:local" "pachyderm/pachd:${version}"
    docker push "pachyderm/pachd:${version}"
    docker tag "pachyderm/worker:local" "pachyderm/worker:${version}"
    docker push "pachyderm/worker:${version}"

    # Avoid having to rebuild unchanged docker image layers every time
    docker save -o docker_images/images.tar $(docker images -a -q)
    ls -alh ${HOME}/docker_images/

    # Push pipeline build images
    make docker-push-pipeline-build
fi

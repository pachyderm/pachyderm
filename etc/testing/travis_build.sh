#!/bin/bash

set -ex

# We can't run the build step if there's no access to the secret env vars
if [[ "$TRAVIS_SECURE_ENV_VARS" == "true" ]]; then
    docker login -u pachydermbuildbot -p "${DOCKER_PWD}"

    # Load saved docker image layers, to avoid having to repeatedly apt-get
    # stuff every time we build
    echo "Loading docker images from cache"
    time docker load -i ${HOME}/docker_images/images.tar || true

    time make install docker-build
    version=$(pachctl version --client-only)
    docker tag "pachyderm/pachd:local" "pachyderm/pachd:${version}"
    docker push "pachyderm/pachd:${version}"
    docker tag "pachyderm/worker:local" "pachyderm/worker:${version}"
    docker push "pachyderm/worker:${version}"

    # Avoid having to rebuild unchanged docker image layers every time
    echo "Saving these docker images to cache:"
    docker images pachyderm/worker
    docker images pachyderm/pachd
    docker images pachyderm_build
    time docker save -o ${HOME}/docker_images/images.tar $(
        docker images -q pachyderm/worker
        docker images -q pachyderm/pachd
        docker images -q pachyderm_build
    )
    ls -alh ${HOME}/docker_images/

    # Push pipeline build images
    make docker-push-pipeline-build
fi

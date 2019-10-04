#!/bin/bash

set -ex

if [[ "$TRAVIS_PULL_REQUEST" != "false" && "$TRAVIS_SECURE_ENV_VARS" == "true" && "$BUCKET" == "admin" ]]; then
    if [[ -z "${DOCKER_USERNAME}" || -z "${DOCKER_PASSWORD}" ]]; then
        echo 'the docker login-related environment variables ($DOCKER_USERNAME and $DOCKER_PASSWORD) were not found; skipping push.'
        exit 0
    fi

    echo "$DOCKER_PASSWORD" | docker login -u "$DOCKER_USERNAME" --password-stdin
    ./etc/build/release_version
    export VERSION="$(shell cat VERSION)"
    docker tag pachyderm/pachd:latest pachyderm/pachd:$VERSION
    docker push pachyderm/pachd:$VERSION
    docker tag pachyderm/worker:latest pachyderm/worker:$VERSION
    docker push pachyderm/worker:$VERSION
fi

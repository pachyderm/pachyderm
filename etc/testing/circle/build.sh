#!/bin/bash

set -ex

## pachctl build
go version
make install
VERSION=$(pachctl version --client-only)

## local docker build

# first, wait for minikube's docker to be ready (minikube is starting up async in a background job)
for _ in $(seq 36); do
    if minikube docker-env &>/dev/null; then
        echo 'minikube docker ready' | ts
        eval "$(minikube docker-env)"
        break
    fi
    echo 'sleeping' | ts
    sleep 5
done

if [[ -z "${MINIKUBE_ACTIVE_DOCKERD}" ]]; then
    echo 'never setup docker, aborting' | ts
    exit 1
fi

# then actually do the build
git config user.email "donotreply@pachyderm.com"
git config user.name "anonymous"
git tag -f -am "Circle CI test v$VERSION" v"$VERSION"
#make docker-build
docker build --network=host -f etc/test-images/Dockerfile.netcat -t pachyderm/ubuntuplusnetcat:local .

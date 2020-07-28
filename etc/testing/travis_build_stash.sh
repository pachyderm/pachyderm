#!/bin/bash

set -ex

# See travis_build_check.sh for context.

# Yay, we got this far (travis_build_check.sh didn't exit 1) so we got a good
# git state. Save it! Our tests will need it too, and might not be able to get
# it back then (or might not get it back with the same commit ID at HEAD, which
# will break trying to pull docker images tagged with the output of pachctl
# version).

if [[ ! "$TRAVIS_SECURE_ENV_VARS" == "true" ]]; then
    echo "Not trying to push pachyderm.tar to docker hub, as we're running for"
    echo "an external contribution."
    exit 0
fi

docker login -u pachydermbuildbot -p "${DOCKER_PWD}"

cd /home/travis/gopath/src/github.com/pachyderm
mkdir -p /tmp/save_git_tarball
tar cf /tmp/save_git_tarball/pachyderm.tar pachyderm
cd /tmp/save_git_tarball

cat <<EOT >Dockerfile
FROM ubuntu:xenial
COPY pachyderm.tar /
EOT

docker build -t pachyderm/ci_code_bundle:"${TRAVIS_BUILD_NUMBER}" .
docker push pachyderm/ci_code_bundle:"${TRAVIS_BUILD_NUMBER}"

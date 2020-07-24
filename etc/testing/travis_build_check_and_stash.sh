#!/bin/bash

set -ex

# Travis doesn't reliably give us the right version of the code in PR builds.
#
# Sometimes it's too old (the GitHub merge ref API can be out of date) and so
# we get code from a previous commit in the pull request.
#
# Sometimes it's too new (a new commit was pushed to the PR before we ran, and
# GitHub updated the merge ref API already).
#
# We know that we are trying to test code in a merge preview with
# $TRAVIS_PULL_REQUEST_SHA on the RHS of the merge though, that information is
# reliable at least.
#
# So, try a few times to fetch and identify a valid version of the code. If we
# time out, give up. If we succeed, stash that code in a docker image which
# later tests can use to get it back.

cd /home/travis/gopath/src/github.com/pachyderm/pachyderm

if [ "${TRAVIS_PULL_REQUEST}" == "false" ]; then
    # These shenannigans not needed for release and branch builds, hopefully.
    exit 0
fi

echo "Detected PR build, checking for consistency..."

tries=0
while true; do
    # Check if HEAD is a merge commit with the commit we want to test as one of
    # the parents (the other one would be current main branch HEAD)
    parents=$(git rev-list --parents -n 1 HEAD)
    if [[ "$parents" == *"$TRAVIS_PULL_REQUEST_SHA"* ]]; then
        echo "Great, found the commit we're meant to be testing ($TRAVIS_PULL_REQUEST_SHA)"
        echo "as one of the parents of the HEAD merge preview commit ($parents)"
        break
    else
        tries=$((tries+1))
        if [ "$tries" -gt 60 ]; then
            echo "Gave up waiting for GitHub to give us the right commit";
            exit 1
        fi
        # The commit we want to test wasn't found yet, maybe the GitHub API is out
        # of date. Let's wait a second and try to update it, then check again.
        sleep 1
        echo "GitHub didn't give us the commit we're meant to be testing ($TRAVIS_PULL_REQUEST_SHA)"
        echo "as one of the parents of the HEAD merge preview commit ($parents). Trying again..."
        git fetch origin +refs/pull/${TRAVIS_PULL_REQUEST}/merge
        git checkout -qf FETCH_HEAD
    fi
done

# Yay, we got a good git state. Save it! Our tests will need it too, and might
# not be able to get it back then (or might not get it back with the same
# commit ID at HEAD, which will break trying to pull docker images tagged with
# the output of pachctl version).

if [[ ! "$TRAVIS_SECURE_ENV_VARS" == "true" ]]; then
    echo "Need travis env vars so we can auth to docker hub."
    exit 1
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

docker build -t pachyderm/ci_code_bundle:${TRAVIS_BUILD_NUMBER} .
docker push pachyderm/ci_code_bundle:${TRAVIS_BUILD_NUMBER}

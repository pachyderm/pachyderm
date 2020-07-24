#!/bin/bash

set -ex

# See travis_build_check_and_stash.sh for context.
#
# This script takes the docker image which that script pushed, and unpacks the
# git source tree from it, as it is more reliably going to be the correct
# source tree for this build than what Travis and GitHub give us (in the PR
# build case, at least).

if [ "${TRAVIS_PULL_REQUEST}" == "false" ]; then
    # These shenannigans not needed for release and branch builds, hopefully.
    exit 0
fi

cd /home/travis/gopath/src/github.com/pachyderm/
mv pachyderm pachyderm.old

docker run -v $(pwd):/unpack pachyderm/ci_code_bundle:${TRAVIS_BUILD_NUMBER} \
    tar xf /pachyderm.tar /unpack/

ls -alh pachyderm
sudo chown -R "${USER}:${USER}" pachyderm

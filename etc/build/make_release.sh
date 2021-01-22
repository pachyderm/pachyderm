#!/bin/bash

set -e

GOPATH=$(go env GOPATH)

make VERSION_ADDITIONAL="$VERSION_ADDITIONAL" install-clean
version="$("$GOPATH/bin/pachctl" version --client-only)"

echo "--- Releasing Version: $version"
make VERSION="$version" VERSION_ADDITIONAL="$VERSION_ADDITIONAL" release

echo "--- Updating homebrew formula to use binaries at version $version"
BRANCH=master
if [ -n "$1" ]
then
    BRANCH=$version
fi

MAJOR_MINOR=$(echo "$version" | cut -f -2 -d ".")

rm -rf homebrew-tap || true
git clone git@github.com:pachyderm/homebrew-tap

pushd homebrew-tap
    git checkout -b "$BRANCH" || git checkout "$BRANCH"
    VERSION=$version ./update-formula.sh
    git add "pachctl@$MAJOR_MINOR.rb"
    git commit -a -m "[Automated] Update formula to release version $version"
    git pull origin "$BRANCH" || true
    git push origin "$BRANCH"
popd

rm -rf homebrew-tap

echo "--- Successfully updated homebrew for version $version ---"

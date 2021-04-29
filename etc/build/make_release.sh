#!/bin/bash

set -e

SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"
# shellcheck source=./../govars.sh
source "${SCRIPT_DIR}/../govars.sh"

make VERSION_ADDITIONAL="$VERSION_ADDITIONAL" install-clean
version="$("${PACHCTL}" version --client-only)"

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

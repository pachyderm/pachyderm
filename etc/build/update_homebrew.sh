#!/bin/bash

set -e

version=$1
stable=$2

echo "--- Updating homebrew formula to use binaries at version $version"
BRANCH=master
if [[ $stable == 0 ]];
then
    BRANCH=$version
fi;

MAJOR_MINOR=$(echo "$version" | cut -f -2 -d ".")

rm -rf homebrew-tap || true
git clone git@github.com:pachyderm/homebrew-tap

pushd homebrew-tap
    git checkout -b "$BRANCH" || git checkout "$BRANCH"
    VERSION=$version ./update-formula.sh
    git add "pachctl@$MAJOR_MINOR.rb"
    if [[ $stable == 1 ]]; then
        cp "pachctl@$MAJOR_MINOR.rb" pachctl.rb
        sed -i -E 's/class PachctlAT\([0-9]*\)/class Pachctl/g' pachctl.rb
        git add pachctl.rb
    fi;
    git commit -a -m "[Automated] Update formula to release version $version"
    git pull origin "$BRANCH" || true
    git push origin "$BRANCH"
popd

rm -rf homebrew-tap

echo "--- Successfully updated homebrew for version $version ---"

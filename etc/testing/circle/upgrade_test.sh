#!/bin/sh

set -ve


git checkout $1 &&

./etc/testing/circle/build.sh &&
./etc/testing/circle/launch.sh &&

git checkout $2 &&
./etc/testing/circle/pre_upgrade_test.sh &&

# build and upgrade to HEAD
./etc/testing/circle/build.sh &&
./etc/testing/circle/helm-upgrade.sh &&
./etc/testing/circle/post_upgrade_test.sh
#!/bin/sh

set -ve


git checkout $1 &&

./etc/testing/circle/build.sh &&
./etc/testing/circle/launch.sh &&

git checkout $2 &&
go test -v ./src/testing/upgrade -run TestPreUpgrade &&

# build and upgrade to HEAD
./etc/testing/circle/build.sh &&
./etc/testing/circle/helm-upgrade.sh &&
go test -v ./src/testing/upgrade -run TestPostUpgrade
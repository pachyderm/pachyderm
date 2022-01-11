#!/bin/sh

set -ve

./etc/testing/circle/build.sh &&

./etc/testing/circle/launch-published.sh &&

go test -v ./src/testing/upgrade -run TestPreUpgrade &&

./etc/testing/circle/helm-upgrade.sh &&

go test -v ./src/testing/upgrade -run TestPostUpgrade
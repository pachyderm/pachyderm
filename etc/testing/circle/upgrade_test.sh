#!/bin/sh

set -ve

./etc/testing/circle/build.sh &&

./etc/testing/circle/launch-published.sh &&

go test -v ./src/testing/upgrade -run TestPreUpgrade -tags=livek8s &&

./etc/testing/circle/helm-upgrade.sh &&

go test -v ./src/testing/upgrade -run TestPostUpgrade -tags=livek8s

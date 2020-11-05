#!/bin/bash

set -e

make VERSION_ADDITIONAL="$VERSION_ADDITIONAL" install-clean
version="$("$GOPATH/bin/pachctl" version --client-only)"

echo "--- Releasing Version: $version"
make VERSION="$version" VERSION_ADDITIONAL="$VERSION_ADDTIONAL" release

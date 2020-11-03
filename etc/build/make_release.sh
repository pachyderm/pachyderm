#!/bin/bash

set -e

if [ -v "$VERSION_ADDITIONAL" ]
then
    echo "Need to specific VERSION_ADDITIONAL! Aborting release"
    exit 1
fi

make VERSION_ADDITIONAL=$VERSION_ADDITIONAL install-clean
version="$("$GOPATH/bin/pachctl" version --client-only)"

echo "--- Releasing Version: $version"
make VERSION=$version VERSION_ADDITIONAL=$VERSION_ADDTIONAL release

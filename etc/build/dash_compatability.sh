#!/bin/bash

VERSION=$("$GOPATH/bin/pachctl" version --client-only)
touch "etc/compatibility/$VERSION"
go run etc/build/get_dash_version.go >> "etc/compatibility/$VERSION"

echo "--- Updated dash compatability file for pachctl $VERSION"


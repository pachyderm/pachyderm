#!/bin/bash

if [ ! -f "$GOPATH/bin/pachctl" ]
then
    echo "$GOPATH/bin/pachctl not found."
    exit 1
fi

RELVERSION=$("$GOPATH/bin/pachctl" version --client-only)
VERSION=$(echo "$RELVERSION" | cut -f -1 -d "-")

touch "etc/compatibility/$VERSION"
go run etc/build/get_dash_version.go >> "etc/compatibility/$VERSION"

# Update the link to latest only for major/minor/patch release
if [ $RELVERSION == $VERSION ]
then
    echo "Update latest --> $RELVERSION"
    ln -s -f "etc/compatibility/$VERSION" "etc/compatibility/latest"
fi

echo "--- Updated dash compatibility file for pachctl $VERSION"


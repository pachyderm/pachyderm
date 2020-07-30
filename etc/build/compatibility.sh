#!/bin/bash

if [ ! -f "$GOPATH/bin/pachctl" ]
then
    echo "$GOPATH/bin/pachctl not found."
    exit 1
fi

RELVERSION=$("$GOPATH/bin/pachctl" version --client-only)
VERSION=$(echo "$RELVERSION" | cut -f -1 -d "-")

go run etc/build/get_default_version.go defaultDashVersion >> "etc/compatibility/$VERSION"
go run etc/build/get_default_version.go defaultIDEVersion >> "etc/compatibility/ide/$VERSION"
go run etc/build/get_default_version.go defaultIDEChartVersion >> "etc/compatibility/jupyterhub/$VERSION"

# Update the link to latest only for major/minor/patch release
if [ "$RELVERSION" == "$VERSION" ]
then
    echo "Update latest --> $RELVERSION"
    ln -s -f "$VERSION" "etc/compatibility/latest"
    ln -s -f "$VERSION" "etc/compatibility/ide/latest"
    ln -s -f "$VERSION" "etc/compatibility/jupyterhub/latest"
fi

echo "--- Updated compatibility files for pachctl $VERSION"

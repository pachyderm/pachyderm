#!/bin/bash

set -e

SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"
source "${SCRIPT_DIR}/../govars.sh"

if [ ! -f "${PACHCTL}" ]
then
    echo "${PACHCTL} not found."
    exit 1
fi

RELVERSION=$("${PACHCTL}" version --client-only)
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

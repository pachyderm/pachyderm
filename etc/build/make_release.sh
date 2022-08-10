#!/bin/bash

set -e

SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"
# shellcheck disable=SC1090
source "${SCRIPT_DIR}/../govars.sh"

INSTALLED_GOVER="$(go version | cut -d ' ' -f 3)"
EXPECTED_GOVER="$(head -3 <"${SCRIPT_DIR}"/../../go.mod | tail -1 | cut -d' ' -f2)"
if [ "${INSTALLED_GOVER}" != "${EXPECTED_GOVER}" ]
then
    echo "Current go version ${INSTALLED_GOVER}"
    echo "Expected go version ${EXPECTED_GOVER}"
    echo "Install the expected version of go before doing a release!"
    exit 1
fi

make VERSION_ADDITIONAL="$VERSION_ADDITIONAL" install-clean
version="$("${PACHCTL}" version --client-only)"

echo "--- Releasing Version: $version"
make VERSION="$version" VERSION_ADDITIONAL="$VERSION_ADDITIONAL" release

./update_homebrew.sh "$version" "$1"

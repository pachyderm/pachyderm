#!/bin/bash
# This script detects any changes to generated protobuf code

set -ex
command -v sha256sum

# cd to top-level pachyderm directory
scriptdir="$(dirname "${0}")"
cd "${scriptdir}/../.."

# hash our generated protobuf code, mostly to see if make proto changed anything
orig_hash="$(
find src -regex ".*\.pb\.go" \
  | sort -u \
  | xargs cat \
  | sha256sum \
  | awk '{print $1}'
)"

make proto

# hash newly-generated code
new_hash="$(
find src -regex ".*\.pb\.go" \
  | sort -u \
  | xargs cat \
  | sha256sum \
  | awk '{print $1}'
)"

# Exit with error if code changed
if test "${orig_hash}" != "${new_hash}"; then
  echo "Protos need to be recompiled; run 'DOCKER_BUILD_FLAGS=--no-cache make proto'."
  exit 1
fi

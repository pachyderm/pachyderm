#!/usr/bin/env bash

set -e

GIT_REPO_DIR=$(cd "$( dirname "${BASH_SOURCE[0]}" )" 2>&1 > /dev/null && git rev-parse --show-toplevel)
cd "${GIT_REPO_DIR}"

# Collect untracked args into '\( -o -path abc -o -path def -o ... \) -prune'
# argument to be passed to 'find'
skip_paths=()
for file in $(git status --porcelain | grep '^??' | sed 's/^?? //'); do
  # Convert to relative path, as required by -path
  file="./$(realpath --relative-to="." "${file}")"
  # 1. Always prefix with -o (i.e. don't fencepost; find always has an argument
  #    before skip_paths, so including the initial -o works whether skip_paths
  #    is empty or not
  # 2. strip trailing slash as required by -path
  skip_paths+=( -o -path "${file%/}" )
done

curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s v1.40.1
./bin/golangci-lint run --timeout 10m

# shellcheck disable=SC2046
find . \
  \( -path ./etc/plugin "${skip_paths[@]}" \) -prune -o -name "*.sh" -print0 \
| xargs -P 16 shellcheck -e SC1091 -e SC2010 -e SC2181 -e SC2004 -e SC2219

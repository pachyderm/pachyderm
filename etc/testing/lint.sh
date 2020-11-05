#!/usr/bin/env bash

set -e

GIT_REPO_DIR=$(cd "$( dirname "${BASH_SOURCE[0]}" )" 2>&1 > /dev/null && git rev-parse --show-toplevel)
cd "${GIT_REPO_DIR}"

# Collect untracked args into '\( -path abc -o -path def -o ... \) -prune'
# argument to be passed to 'find'
skip_paths=()
for file in $(git status --porcelain | grep '^??' | sed 's/^?? //'); do
  file="./$(realpath --relative-to="." "${file}")"
  if [[ "${#skip_paths[@]}" -gt 0 ]]; then
    skip_paths+=( -o )
  fi
  skip_paths+=( -path "${file%/}" ) # strip trailing slash as req'd by -path
done

go get -u golang.org/x/lint/golint
find "./src" \
  \( -path "*.pb.go" "${skip_paths[@]}" \) -prune -o -name '*.go' -print \
| while read -r file; do
    if [[ "${file}" == *pkg/tar* ]]; then
        continue
    fi
    golint -set_exit_status "$file";
done;

files=$(gofmt -l "${GIT_REPO_DIR}/src" || true)
if [[ -n "${files}" ]]; then
    echo Files not passing gofmt:
    tr ' ' '\n'  <<< "$files"
    exit 1
fi

go get honnef.co/go/tools/cmd/staticcheck
staticcheck "${GIT_REPO_DIR}/..."

# shellcheck disable=SC2046
find . \
  \( -path ./etc/plugin "${skip_paths[@]}" \) -prune -o -name "*.sh" -print \
| while read -r file; do
    shellcheck -e SC1091 -e SC2010 -e SC2181 -e SC2004 -e SC2219 "${file}"
done

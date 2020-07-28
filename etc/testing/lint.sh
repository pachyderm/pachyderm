#!/usr/bin/env bash

set -e

GIT_REPO_DIR=$(cd "$( dirname "${BASH_SOURCE[0]}" )" 2>&1 > /dev/null && git rev-parse --show-toplevel)

echo "============================="
echo "Running build & lint"
echo "pwd: $(pwd)"
echo "GIT_REPO_DIR: $GIT_REPO_DIR"
echo "============================="

# go build will install go modules as needed. golint, however, won't, and will complain:
# -: could not analyze dependency github.com/pachyderm/pachyderm/src/client [github.com/pachyderm/pachyderm/src/client.test] of github.com/pachyderm/pachyderm/src/client.test (compile)

go build ./...

go get -u golang.org/x/lint/golint
for file in $(find "${GIT_REPO_DIR}/src" -name '*.go' | grep -v '\.pb\.go'); do
    if [[ $file == *fileset/tar* ]]; then
        continue
    fi
    golint -set_exit_status "$file";
done;

files=$(gofmt -l "${GIT_REPO_DIR}/src" || true)

if [[ $files ]]; then
    echo Files not passing gofmt:
    tr ' ' '\n'  <<< "$files"
    exit 1
fi

go get honnef.co/go/tools/cmd/staticcheck
staticcheck "${GIT_REPO_DIR}/..."
# shellcheck disable=SC2046
shellcheck -e SC2010 -e SC2181 -e SC2004 -e SC2219 $(find . -path ./etc/plugin -prune -o -name "*.sh" -print)

echo "============================="
echo "Lint passed."
echo "============================="

#!/usr/bin/env bash

set -e

GIT_REPO_DIR=$(cd "$( dirname "${BASH_SOURCE[0]}" )" 2>&1 > /dev/null && git rev-parse --show-toplevel)

go get -u golang.org/x/lint/golint
for file in $(find "${GIT_REPO_DIR}/src" -name '*.go' | grep -v '\.pb\.go'); do
    if [[ $file == *fileset/tar* ]]; then
        continue
    fi
    golint -set_exit_status $file;
done;

files=$(gofmt -l ${GIT_REPO_DIR}/src || true)

if [[ $files ]]; then
    echo Files not passing gofmt:
    tr ' ' '\n'  <<< $files
    exit 1
fi

go get honnef.co/go/tools/cmd/staticcheck
staticcheck ${GIT_REPO_DIR}/...

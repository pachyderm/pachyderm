#!/bin/bash

go get -u golang.org/x/lint/golint
for file in $(find "./src" -name '*.go' | grep -v '/vendor/' | grep -v '\.pb\.go'); do \
    golint $file; \
    if [ -n "$(golint $file)" ]; then \
        echo "golint errors!" && echo && exit 1; \
    fi; \
done;

files=$(gofmt -l . | grep -v vendor)

if [[ $files ]]; then
    echo Files not passing gofmt:
    tr ' ' '\n'  <<< $files
    exit 1
fi

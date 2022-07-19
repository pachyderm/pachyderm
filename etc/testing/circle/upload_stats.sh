#!/bin/bash

set -xeuo pipefail

export GOPATH=/home/circleci/.go_workspace
export PATH="${PWD}:${PWD}/cached-deps:${GOPATH}/bin:${PATH}"

if [ -f /tmp/results ]; then
  mkdir -p /tmp/test-results
  go install github.com/jstemmer/go-junit-report@latest
  go-junit-report < /tmp/results > /tmp/test-results/results.xml
fi

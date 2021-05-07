#!/bin/bash

set -xeuo pipefail

export GOPATH=/home/circleci/.go_workspace
export PATH=$(pwd):$(pwd)/cached-deps:$GOPATH/bin:$PATH

if [ -f /tmp/results ]; then
  mkdir -p /tmp/test-results
  go get -u github.com/jstemmer/go-junit-report
  go-junit-report < /tmp/results > /tmp/test-results/results.xml
fi

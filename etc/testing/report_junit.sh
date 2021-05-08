#!/bin/bash

set -xeuo pipefail

GOPATH=/root/go/workspace
export GOPATH
PATH="${GOPATH}/bin:/root/go/bin:${PATH}"
export PATH

if [ -f /tmp/results ]; then
  go get -u github.com/jstemmer/go-junit-report
  go-junit-report < /tmp/results
fi

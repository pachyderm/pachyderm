#!/bin/bash

set -euo pipefail

echo $GOOGLE_TEST_UPLOAD_CREDS > /tmp/google-creds.json
export GOOGLE_APPLICATION_CREDENTIALS=/tmp/google-creds.json

set -x

GOPATH=/root/go/workspace
export GOPATH
PATH="${GOPATH}/bin:/root/go/bin:${PATH}"
export PATH

if [ -f /tmp/results.json ]; then
  go get github.com/actgardner/test-stat
  test-stat /tmp/results.json
fi

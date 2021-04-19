#!/bin/bash

set -euo pipefail

echo $GOOGLE_TEST_UPLOAD_CREDS > /tmp/google-creds.json
export GOOGLE_APPLICATION_CREDENTIALS=/tmp/google-creds.json

set -x

GOPATH=/root/go/workspace
export GOPATH
PATH="${GOPATH}/bin:/root/go/bin:${PATH}"
export PATH

if [ -f /tmp/results ]; then
  curl -L https://github.com/actgardner/test-stat/releases/download/0.1/test-stat-linux-amd64-0.1 -o test-stat
  chmod +x test-stat
  go tool test2json < /tmp/results > /tmp/results.json
  ./test-stat /tmp/results.json
fi

#!/bin/bash

set -euo pipefail

echo $GOOGLE_TEST_UPLOAD_CREDS > /tmp/google-creds.json
gcloud auth activate-service-account --key-file=/tmp/google-creds.json

set -x

GOPATH=/root/go
export GOPATH
PATH="${GOPATH}/bin:${PATH}"
export PATH

go install github.com/actgardner/test-stat
test-stat /tmp/results.json

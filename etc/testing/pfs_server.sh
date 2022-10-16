#!/bin/bash
set -euxo pipefail

export TIMEOUT=$1

go test -v -count=1 -tags=k8s ./src/server/pfs/server -timeout "$TIMEOUT"

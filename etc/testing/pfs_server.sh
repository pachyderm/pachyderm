#!/bin/bash
set -euxo pipefail

export TIMEOUT=$1

go test -v -count=1 ./src/server/pfs/server ./src/server/pfs/server/testing -timeout "$TIMEOUT" -tags=k8s

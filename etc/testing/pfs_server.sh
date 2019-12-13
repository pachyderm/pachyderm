#!/bin/bash
set -euxo pipefail

export TIMEOUT=$1

# grep out PFS server logs, as otherwise the test output is too verbose to
# follow and breaks travis
go test -v -count=1 ./src/server/pfs/server ./src/server/pfs/server/testing -timeout $TIMEOUT | grep -v "$(date +^%FT)"

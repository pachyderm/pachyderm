#!/bin/sh
# run this from the root

set -ve

# internal
mkdir -p ./src/internal
git mv ./src/server/pkg/* ./src/internal
git mv ./src/client/pkg/* ./src/internal

# public api
git mv ./src/client/pfs ./src/pfs
git mv ./src/client/pps ./src/pps
git mv ./src/client/auth ./src/auth
git mv ./src/client/admin ./src/admin
git mv ./src/client/debug ./src/debug
git mv ./src/client/enterprise ./src/enterprise
git mv ./src/client/health ./src/health
git mv ./src/client/transaction/ ./src/transaction

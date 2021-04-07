#!/bin/bash -euo pipefail

go get github.com/vektra/mockery/v2/.../
rm -rf src/internal/mocks
mkdir -p src/internal/mocks

for f in $(find src -maxdepth 1 -mindepth 1 -type d | grep -v 'server' | grep -v 'internal' | grep -v 'client'); do
    mockery --output src/internal/mocks/$(basename $f) --keeptree --all --dir $f
done

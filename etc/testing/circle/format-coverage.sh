#!/bin/bash
set -euxo pipefail

find "$TEST_RESULTS" -name "*.tgz" -type f -exec tar xzf {} --transform="s/.*\///" -C "$TEST_RESULTS" \; # extract all covdata to test-results folder
go tool covdata textfmt -i="$TEST_RESULTS" -o "$TEST_RESULTS/coverage.txt"
find . -name "*covcounters.*" -type f -exec rm {} \; # cleanup so we don't store tons of extra files in CI artifacts
find . -name "*covmeta.*" -type f -exec rm {} \; 
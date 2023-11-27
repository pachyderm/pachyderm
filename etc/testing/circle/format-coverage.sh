#!/bin/bash
set -euxo pipefail

ls "$TEST_RESULTS" 
find "$TEST_RESULTS" -name "*.tgz" -type f -exec tar xvf {} --transform="s/.*\///" -C "$TEST_RESULTS" \; # extract all covdata to test-results folder
go tool covdata textfmt -i="$TEST_RESULTS" -o "$TEST_RESULTS/coverage.txt"
find "$TEST_RESULTS" -name "*covcounters.*" -type f -exec rm {} \; # cleanup so we don't store tons of extra files in CI artifacts
find "$TEST_RESULTS" -name "*covmeta.*" -type f -exec rm {} \; 
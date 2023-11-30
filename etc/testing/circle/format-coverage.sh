#!/bin/bash
set -euxo pipefail

ls "$TEST_RESULTS" 
find "$TEST_RESULTS" -name "*.tgz" -type f -exec tar xvf {} --transform="s/.*\///" -C "$TEST_RESULTS" \; # extract all covdata to test-results folder
ls "$TEST_RESULTS"
ls "$TEST_RESULTS/cover"
go tool covdata textfmt -i="$TEST_RESULTS/cover" -o "$TEST_RESULTS/coverage.txt"
find "$TEST_RESULTS/cover" -name "*covcounters.*" -type f -exec rm {} \; # cleanup so we don't store tons of extra files in CI artifacts
find "$TEST_RESULTS/cover" -name "*covmeta.*" -type f -exec rm {} \; 
#!/bin/bash
set -euxo pipefail

cover_folder="${TEST_RESULTS}/cover" 
find "$cover_folder" -name "*.tgz" -type f -exec tar xvf {} --transform="s/.*\///" -C "$TEST_RESULTS" \; # extract all covdata to test-results folder
go tool covdata textfmt -i="$cover_folder" -o "$TEST_RESULTS/coverage.txt"
find "$cover_folder" -name "*covcounters.*" -type f -exec rm {} \; # cleanup so we don't store tons of extra files in CI artifacts
find "$cover_folder" -name "*covmeta.*" -type f -exec rm {} \; 
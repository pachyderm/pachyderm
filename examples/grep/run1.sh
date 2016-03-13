#!/bin/sh

set -xEe

pachctl create-repo data
commit1="$(pachctl start-commit data)"
cat examples/grep/set1.txt | pachctl put-file data "$commit1" set1.txt
pachctl finish-commit data "$commit1"
pachctl create-pipeline -f examples/grep/grep.json
pachctl create-pipeline -f examples/grep/count.json

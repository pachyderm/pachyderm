#!/bin/sh

set -xEe

# create the first data commit
pachctl create-repo data
commit1="$(pachctl start-commit data)"
cat examples/grep/set1.txt | pachctl put-file data "$commit1" set1.txt
pachctl finish-commit data "$commit1"

# create the pipelines
pachctl create-pipeline -f examples/grep/grep.json
pachctl create-pipeline -f examples/grep/count.json

# create the secont data commit
commit2="$(pachctl start-commit data $commit1)"
cat examples/grep/set2.txt | pachctl put-file data "$commit2" set2.txt
pachctl finish-commit data "$commit2"

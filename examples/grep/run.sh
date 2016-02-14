#!/bin/sh

set -xEe

pachctl create-repo data
commit1="$(pachctl start-commit data)"
cat examples/grep/set1.txt >/pfs/data/"$commit1"/set1.txt
pachctl finish-commit data "$commit1"
pachctl create-pipeline -f examples/grep/grep.json
pachctl create-pipeline -f examples/grep/count.json
commit2="$(pachctl start-commit data $commit1)"
cat examples/grep/set2.txt >/pfs/data/"$commit2"/set2.txt
pachctl finish-commit data "$commit2"

#!/bin/sh

set -xEe

commit2="$(pachctl start-commit data $commit1)"
cat examples/grep/set2.txt | pachctl put-file data "$commit2" set2.txt
pachctl finish-commit data "$commit2"

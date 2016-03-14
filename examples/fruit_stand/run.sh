#!/bin/sh

set -xEe

# create the first data commit
pachctl create-repo data
pachctl start-commit data master
cat examples/fruit_stand/set1.txt | pachctl put-file data master sales
pachctl finish-commit data master
# create the pipelines
pachctl create-pipeline -f examples/fruit_stand/pipeline.json
# create the secont data commit
pachctl start-commit data master
cat examples/fruit_stand/set2.txt | pachctl put-file data master sales
pachctl finish-commit data master

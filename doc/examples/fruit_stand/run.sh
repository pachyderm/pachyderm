#!/bin/sh

set -xEe

THIS_DIR="$(dirname $0)"

# create the first data commit
pachctl create-repo data
pachctl start-commit data master
cat ${THIS_DIR}/set1.txt | pachctl put-file data master sales
pachctl finish-commit data master
echo "wrote first commit to data"
pachctl list-file data master
# create the pipelines
pachctl create-pipeline -f doc/examples/fruit_stand/pipeline.json
echo "created pipeline"
# create the second data commit
pachctl start-commit data master
cat ${THIS_DIR}/set2.txt | pachctl put-file data master sales
pachctl finish-commit data master
echo "wrote second commit to data"

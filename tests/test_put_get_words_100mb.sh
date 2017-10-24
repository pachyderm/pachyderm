#!/bin/bash

pachctl version || {
  echo "Could not connect to Pachyderm"
  exit 1
}

set -x

THIS_DIR="$(dirname $0)"

############
# Delete all repos from previous experiments
############
[[ -d ./words_output ]] && rm -r ./words_output
pachctl delete-repo words_test
pachctl create-repo words_test
mkdir ./words_output

############
# Get data into pachyderm via GCS bucket path
############
COMMIT=$(pachctl start-commit words_test master | tee /dev/stderr)
pachctl put-file words_test "${COMMIT}" / -r -f gs://words_testdata
pachctl finish-commit words_test "${COMMIT}"

############
# Get data back out of pachyderm
############
pachctl get-file words_test master /data_100mb.txt -o words_output/data_100mb.txt

./testdata/words/diff.py ./testdata/words/data_100mb.txt ./words_output/data_100mb.txt

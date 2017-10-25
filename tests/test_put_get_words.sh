#!/bin/bash

pachctl version || {
  echo "Could not connect to Pachyderm"
  exit 1
}

############
# Parse flags
############
FILE=data_100mb.txt
eval "set -- $( getopt -l "file:" -- "${0}" "${@}" )"
while true; do
    case "${1}" in
        --file)
          FILE="${2}"
          shift 2
          ;;
        --)
          shift
          break
          ;;
    esac
done

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
# pachctl put-file words_test "${COMMIT}" / -r -f gs://words_testdata
pachctl put-file words_test "${COMMIT}" /${FILE} -f gs://words_testdata/${FILE}
pachctl finish-commit words_test "${COMMIT}"
pachctl list-file words_test "${COMMIT}" /

############
# Get data back out of pachyderm
############
pachctl get-file words_test master /${FILE} -o words_output/${FILE}

./testdata/words/diff.py ./testdata/words/${FILE} ./words_output/${FILE}

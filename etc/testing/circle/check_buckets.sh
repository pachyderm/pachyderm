#!/bin/bash
set -e
IFS=

function fail {
    echo "$1"
    exit 1
}

# child field_name finds a the first top-level YAML field named field_name in the input
# and sends through any lines under that field, stripped of extra leading space
function child {
    target="^(- )?$1:"
    prefix=0

    while read line; do
        if [[ $prefix -gt 0 ]]; then
            if [[ $line =~ ^[[:graph:]] ]]; then
                exit 0
            fi
            echo ${line:$prefix}
        fi

        if [[ $line =~ $target ]]; then
            if [[ $line =~ ^- ]]; then
                prefix=4
            else
                prefix=2
            fi
        fi
    done
}

count=$(cat .circleci/config.yml | child jobs | child circleci | child environment | grep TEST_BUCKETS | cut -d \" -f 2)

echo "should be $count buckets, checking for TEST$count"

cat .circleci/config.yml | child workflows | child circleci | child jobs \
    | child circleci | child matrix | child parameters | grep "TEST$count" || fail "test bucket number mismatch"


#!/bin/bash

set -e

if [ $# -ne 1 ]
then
    echo "Need to specify change log file used in goreleaser"
    exit 1
fi

git diff HEAD^ HEAD -- CHANGELOG.md | sed -n /##/,/##/p | grep -v "##" | cut -d '+' -f 2 > $1

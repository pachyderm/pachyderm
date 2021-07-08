#!/bin/bash

# Since we use our own changelog, go releaser expects the release notes to be passed in for 
# just the current release. This script expects the last commit to be the updated release notes
# and outputs the notes for the latest release to the given file. Notes must be formatted by
# delintating a new release with a markdown h2 tag (##)

set -e

if [ $# -ne 1 ]
then
    echo "Need to specify change log file used in goreleaser"
    exit 1
fi

mkdir -p "$(dirname "$1")" && touch "$1"
git diff HEAD^ HEAD -- CHANGELOG.md | sed -n /##/,/##/p | grep -v "##" | cut -d '+' -f 2 > "$1"

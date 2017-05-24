#!/usr/bin/env bash

# Disable -composites because we don't consider those real issues.
# Disable -printf because we unfortunately stored the output of an
# incorrect Sprintf into etcd, which means that if we fix it, we
# will need a migration.
OUTPUT=$(go tool vet -composites=false -printf=false . |& grep -v /vendor/)

echo $OUTPUT

if [[ -n $OUTPUT ]]
then
    echo $OUTPUT
    exit 1
fi

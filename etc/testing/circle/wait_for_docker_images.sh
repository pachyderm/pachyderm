#!/bin/bash

set -euxo pipefail

for _ in $(seq 1 120); do
    if docker manifest inspect "pachyderm/pachd:$CIRCLE_SHA1"; then
        echo "manifest exists"
        exit 0
    fi
    echo "sleeping..."
    sleep 5
done

echo "images not available after 600 seconds; giving up"
exit 1

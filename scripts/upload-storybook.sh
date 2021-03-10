#!/bin/bash

set -euo pipefail
IFS=$'\n\t'

COMPONENT_LIBARY_CHANGED=$(git diff --name-status HEAD~1...HEAD packages/components)

if [ "$COMPONENT_LIBARY_CHANGED" ]; then
    cd packages/components
    npm ci
    npm run publish:storybook
else
    echo 'Skipped storybook upload'
fi

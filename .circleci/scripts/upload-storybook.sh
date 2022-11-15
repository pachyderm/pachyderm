#!/bin/bash

set -euo pipefail
IFS=$'\n\t'

COMPONENT_LIBARY_CHANGED=$(git diff --name-status HEAD~1...HEAD frontend/components)

if [ "$COMPONENT_LIBARY_CHANGED" ]; then
    cd frontend
    npm ci
    npm run components:publish:storybook
else
    echo 'Skipped storybook upload'
fi

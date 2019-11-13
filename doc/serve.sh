#!/bin/bash

if ! netlify --version; then
  echo "You must install the netlify cli to test our docs build"
  echo "Try running:"
  echo "brew install npm"
  echo "npm install netlify-cli -g"
fi

here="$(dirname "${0}")"
cd "${here}/.."
netlify dev

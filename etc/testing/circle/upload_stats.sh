#!/bin/bash

set -xeuo pipefail

if [ -f /tmp/results ]; then
  mkdir -p /tmp/test-results
  go install github.com/jstemmer/go-junit-report@latest
  go-junit-report < /tmp/results > /tmp/test-results/results.xml
fi

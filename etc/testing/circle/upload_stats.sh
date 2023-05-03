#!/bin/bash

set -xeuo pipefail

if [ -f /tmp/test-results/go-test-results.jsonl ]; then
  mkdir -p /tmp/test-results
  go install github.com/jstemmer/go-junit-report/v2@latest
  go-junit-report -parser gojson < /tmp/test-results/go-test-results.jsonl > /tmp/test-results/results.xml
fi

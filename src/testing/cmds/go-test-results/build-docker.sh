#!/bin/bash
set -euxo pipefail

# Run in go-test-results folder
GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -C ./egress

docker build . -t pachyderm/go-test-results:0.0.9

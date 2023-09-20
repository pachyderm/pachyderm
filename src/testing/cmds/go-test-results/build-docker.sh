#!/bin/bash
set -euxo pipefail

# Run in go-test-results folder
GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -C ./egress
GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -C ./cleanup-cron

docker build . -t pachyderm/go-test-results:0.0.11

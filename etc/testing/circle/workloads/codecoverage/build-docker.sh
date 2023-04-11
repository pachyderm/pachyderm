#!/bin/bash
set -euxo pipefail

# Run in codecoverage folder
GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -C ./collector
GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -C ./extractor
GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -C ./merger

docker build . -t pachyderm/codecoverage:local
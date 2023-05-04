#!/bin/bash
set -euxo pipefail

export TEST_RESULTS='/tmp/test-results'
export CIRCLE_BRANCH='djanicek/core-1553/collect-test-results'
export CIRCLE_JOB='integration-tests-AUTH'
export OPS_PACHD_ADDRESS='grpc://localhost:30650'

pachctl create project ci-metrics
pachctl create repo go-test-results-raw --project ci-metrics

go run -v etc/testing/circle/workloads/go-test-results/collector/main.go

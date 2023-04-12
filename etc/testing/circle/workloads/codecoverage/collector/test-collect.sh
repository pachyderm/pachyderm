#!/bin/bash
set -euxo pipefail

export TEST_RESULTS='/tmp/test-results'
export CIRCLE_BRANCH='djanicek/core-1550/integration-test-cover'
export CIRCLE_JOB='integration-tests-AUTH'
export COV_PACHD_ADDRESS='grpc://localhost:30650'

pachctl create project ci-metrics
pachctl create repo codecoverage-source --project ci-metrics
pachctl create pipeline -f etc/testing/circle/workloads/codecoverage/extractor/extractor.json --project ci-metrics
pachctl create pipeline -f etc/testing/circle/workloads/codecoverage/merger/merger.json --project ci-metrics
pachctl create pipeline -f etc/testing/circle/workloads/codecoverage/uploader/uploader.json --project ci-metrics

# splitting folders simulates the CI workflow better but is uneccessary.
export TEST_RESULTS='/tmp/test-results/amd64 deploy tests'
go run etc/testing/circle/workloads/codecoverage/collector/main.go

export CIRCLE_BRANCH='djanicek/core-1550/integration-test-cover-2'
export TEST_RESULTS='/tmp/test-results/arm64 deploy tests'
go run etc/testing/circle/workloads/codecoverage/collector/main.go

export CIRCLE_BRANCH='djanicek/core-1550/integration-test-cover-3'
export TEST_RESULTS='/tmp/test-results/integration-tests-AUTH'
go run etc/testing/circle/workloads/codecoverage/collector/main.go
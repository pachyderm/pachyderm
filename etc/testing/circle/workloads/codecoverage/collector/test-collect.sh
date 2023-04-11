#!/bin/bash
set -euxo pipefail

# export CIRCLE_WORKFLOW_ID='88a2a33a-44c6-4e8b-bea9-9e5cd88fd1d2'
export CIRCLE_BRANCH='djanicek/core-1550/integration-test-cover'
export CIRCLE_JOB='integration-tests-AUTH'
export COV_PACHD_ADDRESS='grpc://localhost:30650'

pachctl create project ci-metrics
pachctl create repo codecoverage-source --project ci-metrics
pachctl create pipeline -f etc/testing/circle/workloads/codecoverage/extractor/extractor.json --project ci-metrics
go run etc/testing/circle/workloads/codecoverage/collector/main.go

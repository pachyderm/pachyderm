#!/bin/bash
set -euxo pipefail


export TEST_RESULTS='/tmp/test-results'
export CIRCLE_BRANCH='djanicek/core-1553/collect-test-results'
export CIRCLE_JOB='integration-tests-AUTH'
export CIRCLE_WORKFLOW_JOB_ID='51da2066-a881-4591-af6b-fd19061a5c91'
export CIRCLE_WORKFLOW_ID='61da2066-a881-4591-af6b-fd19061a5c92'
export CIRCLE_SHA1='51da2066-a881-4591-af6b-fd19061a5c99'
export CIRCLE_USERNAME='daniel.janicek@hpe.com'
export CIRCLE_NODE_TOTAL=1
export CIRCLE_NODE_INDEX=0
export OPS_PACHD_ADDRESS='grpc://localhost:30650'

pachctl create project ci-metrics
pachctl create repo go-test-results-raw --project ci-metrics
pachctl create pipeline --jsonnet src/testing/cmds/go-test-results/egress/pipeline.jsonnet --arg version=0.0.8 --arg pghost=postgres --arg pguser=pachyderm --project ci-metrics

go run -v src/testing/cmds/go-test-results/collector/main.go

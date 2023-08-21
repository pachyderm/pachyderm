#!/bin/bash
set -euxo pipefail


export TEST_RESULTS='/tmp/test-results'
export CIRCLE_BRANCH='rauh/core-1734-errcommitnotfinished-unavailable-status'
export CIRCLE_JOB='integration-tests-S3_AUTH'
export CIRCLE_WORKFLOW_JOB_ID='51da2066-a881-4591-af6b-fd19061a5c91'
export CIRCLE_WORKFLOW_ID='61da2066-a881-4591-af6b-fd19061a5c92'
export CIRCLE_SHA1='69d1d18b8702d582408dad25987765a433e12b4b' # '238c3336435bace7a0cb71938ad02043933c3289' # '69d1d18b8702d582408dad25987765a433e12b4b' # 99353b4043f5bb542ab63b637f32c609fc9f0a40
export CIRCLE_USERNAME='daniel.janicek@hpe.com'
export CIRCLE_NODE_TOTAL=1
export CIRCLE_NODE_INDEX=0
export OPS_PACHD_ADDRESS='grpc://localhost:30650'

export PIPELINES_VERSION='0.0.10'

pachctl create project ci-metrics
pachctl create repo go-test-results-raw --project ci-metrics
pachctl create pipeline --jsonnet src/testing/cmds/go-test-results/egress/pipeline.jsonnet --arg version="$PIPELINES_VERSION" --arg pghost=postgres --arg pguser=pachyderm --project ci-metrics
pachctl create pipeline --jsonnet src/testing/cmds/go-test-results/cleanup-cron/pipeline.jsonnet --arg version="$PIPELINES_VERSION" --arg pghost=postgres --arg pguser=pachyderm --project ci-metrics

go run -v src/testing/cmds/go-test-results/collector/main.go

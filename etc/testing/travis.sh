#!/bin/bash

set -ex

echo "Running test suite based on BUCKET=$BUCKET"

PPS_SUITE=`echo $BUCKET | grep PPS > /dev/null; echo $?`

make install
make docker-build

kind load docker-image pachyderm/worker:local
kind load docker-image pachyderm/pachd:local

#make docker-build-kafka

make launch-dev

#Misc tests
make lint test-pfs-server test-pfs-cmds test-pfs-storage test-deploy-cmds \
    test-libs test-vault test-auth test-enterprise test-worker test-admin \
    test-s3gateway-integration test-proto-static test-transaction

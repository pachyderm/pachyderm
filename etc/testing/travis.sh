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
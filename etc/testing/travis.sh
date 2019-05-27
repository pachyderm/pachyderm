#!/bin/bash

set -ex

echo "Running test suite based on BUCKET=$BUCKET"

PPS_SUITE=`echo $BUCKET | grep PPS > /dev/null; echo $?`

make install
make docker-build
#COMPILE_IMAGE = "pachyderm/compile:$(shell cat etc/compile/GO_VERSION)"
#make docker-build-worker 

#docker-build-pachd 
#docker-wait-worker 
#docker-wait-pachd

#make docker-build-kafka

make launch-dev
#!/bin/bash

yes | pachctl delete-all
pachctl s3gateway &
pid=$!
outfile=`date '+%Y-%m-%d-%H-%M-%S'`.txt

pushd ./etc/testing/s3gateway/s3-tests
	echo "Running conformance tests -- stderr will be persisted to ./etc/testing/s3gateway/$outfile"
	S3TEST_CONF=../s3gateway.conf ./virtualenv/bin/nosetests 2> >(tee ../$outfile >&2)
popd

kill -SIGINT $pid

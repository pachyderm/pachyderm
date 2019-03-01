#!/bin/bash

yes | pachctl delete-all
pachctl s3gateway &
pid=$!

function finish {
	kill -SIGINT $pid
}
trap finish EXIT

until $(curl --output /dev/null --silent --head --fail http://localhost:30600); do
    echo "Waiting for http://localhost:30600..."
    sleep 1
done

pushd ./etc/testing/s3gateway/s3-tests
	outfile=`date '+%Y-%m-%d-%H-%M-%S'`.txt
	echo "Running conformance tests -- stderr will be persisted to ./etc/testing/s3gateway/$outfile"
	S3TEST_CONF=../s3gateway.conf ./virtualenv/bin/nosetests 2> >(tee ../$outfile >&2)
popd

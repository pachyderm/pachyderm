#!/bin/bash
set -e -o pipefail

export ETCD_IMAGE=$1
export TIMEOUT=$2

# If etcd is not available, start it in a docker container
if ! ETCDCTL_API=3 etcdctl --endpoints=127.0.0.1:2379 get "testkey"; then
    docker run \
        --publish 32379:2379 \
        --ulimit nofile=2048 \
        $ETCD_IMAGE \
        etcd \
        --listen-client-urls=http://0.0.0.0:2379 \
        --advertise-client-urls=http://0.0.0.0:2379 &
fi

# grep out PFS server logs, as otherwise the test output is too verbose to
# follow and breaks travis
go test -v ./src/server/pfs/server -timeout $TIMEOUT | grep -v "$(date +^%FT)"

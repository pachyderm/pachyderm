#!/bin/bash
set -euxo pipefail

export ETCD_IMAGE=$1
export TIMEOUT=$2

# If etcd is not available, start it in a docker container
if ! ETCDCTL_API=3 etcdctl --endpoints=127.0.0.1:32379 get "testkey"; then
    docker run \
        --rm \
        --publish 32379:2379 \
        --ulimit nofile=2048 \
        $ETCD_IMAGE \
        etcd \
        --listen-client-urls=http://0.0.0.0:2379 \
        --advertise-client-urls=http://0.0.0.0:2379 &
fi

# grep out PFS server logs, as otherwise the test output is too verbose to
# follow and breaks travis
go test -v -count=1 ./src/server/pfs/server -timeout $TIMEOUT | grep -v "$(date +^%FT)"

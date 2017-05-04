#!/bin/sh

set -Ee

docker run \
    -d \
    --volume=/:/rootfs:ro \
    --volume=/sys:/sys:ro \
    --volume=/dev:/dev \
    --volume=/var/lib/docker/:/var/lib/docker:rw \
    --volume=/var/lib/kubelet/:/var/lib/kubelet:rw,shared \
    --volume=/var/run:/var/run:rw \
    --net=host \
    --pid=host \
    --privileged=true \
    gcr.io/google_containers/hyperkube:v1.6.2 \
    sh "/rootfs/$PWD/etc/kube/internal.sh"
until kubectl version 2>/dev/null >/dev/null; do sleep 5; done

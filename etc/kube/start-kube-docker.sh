#!/bin/sh

set -Ee

docker run \
    -d \
    --volume=/:/rootfs:ro \
    --volume=/sys:/sys:ro \
    --volume=/dev:/dev \
    --volume=/var/lib/docker/:/var/lib/docker:rw \
    --volume=/var/lib/kubelet/:/var/lib/kubelet:rw \
    --volume=/var/run:/var/run:rw \
    --net=host \
    --pid=host \
    --privileged=true \
    gcr.io/google_containers/hyperkube:v1.2.0 \
    /hyperkube kubelet \
        --containerized \
        --hostname-override="127.0.0.1" \
        --address="0.0.0.0" \
        --api-servers=http://localhost:8080 \
        --config=/etc/kubernetes/manifests \
        --allow-privileged=true
docker run \
    -d \
    --net=host \
    --privileged=true \
    gcr.io/google_containers/hyperkube:v1.2.0 \
    /hyperkube \
    proxy \
    --master=http://127.0.0.1:8080 \
    --v=2
until kubectl version 2>/dev/null >/dev/null; do sleep 5; done

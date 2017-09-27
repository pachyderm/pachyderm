#!/bin/sh

REGISTRY="${1}"

docker load -o pachd
docker tag
docker load -o worker
docker load -o pause
docker load -o etcd
docker load -o dash
docker load -o proxy

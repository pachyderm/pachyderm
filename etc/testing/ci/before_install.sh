#!/bin/bash

set -ex

echo 'DOCKER_OPTS="-H unix:///var/run/docker.sock -s devicemapper"' | tee /etc/default/docker
> /dev/null
apt-get install -qq pkg-config fuse
modprobe fuse
chmod 666 /dev/fuse
cp etc/build/fuse.conf /etc/fuse.conf
chown root:$USER /etc/fuse.conf
mkdir -p /var/lib/kubelet
mount -o bind /var/lib/kubelet /var/lib/kubelet
mount --make-shared /var/lib/kubelet
# Decrypt gke credentials used by batch tests
openssl aes-256-cbc -K $encrypted_1f5361f82e9b_key -iv $encrypted_1f5361f82e9b_iv -in ./etc/testing/pach-travis-86b9f180aa16.json.enc -out ./etc/testing/pach-travis-86b9f180aa16.json -d

#!/bin/sh

set -Ee

truncate /pfs-img/btrfs.img -s 10G
mkfs.btrfs /pfs-img/btrfs.img
while [ 1 ]; do
  sleep 5
done

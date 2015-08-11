#!/bin/sh

set -Ee

mkdir -p /pfs/btrfs
truncate /pfs/btrfs.img -s 10G
mkfs.btrfs /pfs/btrfs.img
mount /pfs/btrfs.img /pfs/btrfs

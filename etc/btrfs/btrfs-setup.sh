#!/bin/sh

set -Ee

truncate /pfs/btrfs.img -s 10G
mkfs.btrfs /pfs/btrfs.img
mount /pfs/btrfs.img /pfs/btrfs

#!/bin/sh

set -Ee

truncate /pfs-img/btrfs.img -s 10G
mkfs.btrfs /pfs-img/btrfs.img

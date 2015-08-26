#!/bin/sh

set -Ee

if [ -z "${BTRFS_DEVICE}" ]; then
  echo "error: BTRFS_DEVICE must be set" >&2
  exit 1
fi

mkdir -p /pfs/btrfs
mount "${BTRFS_DEVICE}" /pfs/btrfs
$@

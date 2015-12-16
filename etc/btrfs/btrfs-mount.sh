#!/bin/sh

set -Ee

if [ -z "${BTRFS_DEVICE}" ]; then
  echo "error: BTRFS_DEVICE must be set" >&2
  exit 1
fi

sudo modprobe loop
mkdir -p $(dirname "${BTRFS_DEVICE}")
truncate "${BTRFS_DEVICE}" -s 10G
mkfs.btrfs "${BTRFS_DEVICE}"
mkdir -p "${PFS_DRIVER_ROOT}"
mount "${BTRFS_DEVICE}" "${PFS_DRIVER_ROOT}"
$@

#!/bin/bash

NVIDIA_RUNNER=$1
cp "$NVIDIA_RUNNER" /rootfs
cp install-nvidia-driver.sh /rootfs

chroot /rootfs ./install-nvidia-driver.sh "${NVIDIA_RUNNER}"

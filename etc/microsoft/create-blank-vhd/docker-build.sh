#!/bin/sh

apk add --update --no-cache \
    git \
    ntfs-3g-progs \
    e2fsprogs \
    qemu-img

# Install go dependencies
go get github.com/docopt/docopt-go
go get github.com/Microsoft/azure-vhd-utils
go install github.com/Microsoft/azure-vhd-utils
go get github.com/Azure/azure-sdk-for-go/storage

#!/bin/sh

apk add --update --no-cache \
    git \
    ntfs-3g-progs \
    e2fsprogs

apk add qemu-img --update-cache --no-cache --repository http://dl-3.alpinelinux.org/alpine/edge/main/ --allow-untrusted

## Install go dependencies
go get github.com/docopt/docopt-go
go get github.com/Microsoft/azure-vhd-utils-for-go
go install github.com/Microsoft/azure-vhd-utils-for-go

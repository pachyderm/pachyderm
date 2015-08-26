FROM ubuntu:14.04
MAINTAINER peter@pachyderm.io

RUN \
  apt-get update -yq && \
  apt-get install -yq --no-install-recommends \
    btrfs-tools
RUN mkdir /pfs-img
VOLUME ["/pfs-img"]
RUN truncate /pfs-img/btrfs.img -s 10G && mkfs.btrfs /pfs-img/btrfs.img

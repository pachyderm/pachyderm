FROM ubuntu:15.04
MAINTAINER peter@pachyderm.io

RUN \
  apt-get update -yq && \
  apt-get install -yq --no-install-recommends \
    btrfs-tools
VOLUME ["/pfs"]
RUN mkdir -p /pfs/btrfs
ADD btrfs-setup.sh /
ENTRYPOINT ["sh", "/btrfs-setup.sh"]

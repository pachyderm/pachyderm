FROM ubuntu:14.04
MAINTAINER peter@pachyderm.io

RUN \
  apt-get update -yq && \
  apt-get install -yq --no-install-recommends \
    btrfs-tools
RUN mkdir /pfs
VOLUME ["/pfs"]
ADD btrfs-setup.sh /
ENTRYPOINT ["sh", "/btrfs-setup.sh"]

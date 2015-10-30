FROM ubuntu:14.04
MAINTAINER jdoliner@pachyderm.io

RUN \
  apt-get update -yq && \
  apt-get install -yq --no-install-recommends \
    btrfs-tools
RUN mkdir /pfs-img
VOLUME ["/pfs-img"]
ADD entrypoint.sh /
ENTRYPOINT ["sh", "entrypoint.sh"]

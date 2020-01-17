FROM ubuntu:16.04
LABEL maintainer="jdoliner@pachyderm.io"

RUN mkdir -p /src/server /etc/testing/artifacts
ADD \
  ./pachyderm_test \
  ./pfs_server_test \
  ./pfs_cmds_test \
  ./pps_cmds_test \
  ./auth_server_test \
  ./auth_cmds_test \
  ./enterprise_server_test \
  ./worker_test \
  ./collection_test \
  ./hashtree_test \
  ./cert_test \
  ./vault_test \
  /src/server/
ADD ./giphy.gif /etc/testing/artifacts
WORKDIR /src/server

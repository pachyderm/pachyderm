FROM ubuntu:16.04
LABEL maintainer="jdoliner@pachyderm.io"

RUN mkdir /pach
ADD ./worker.sh /pach/
ADD ./worker /pach/
ADD ca-certificates.crt /etc/ssl/certs/

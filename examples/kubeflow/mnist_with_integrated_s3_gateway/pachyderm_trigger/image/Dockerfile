FROM ubuntu:19.04

RUN apt-get update -y \
  && apt-get install -y \
  curl \
  jq \
  apt-transport-https \
  ca-certificates gnupg

ADD trigger.sh /bin
RUN chmod +x /bin/trigger.sh

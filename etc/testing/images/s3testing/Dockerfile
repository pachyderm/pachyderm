FROM ubuntu:18.04
LABEL maintainer=msteffen@pachyderm.io

RUN apt-get update -y && apt-get install -y \
  curl \
  python3 \
  python3-dev \
  python3-setuptools \
  groff \
  apt-transport-https \
  ca-certificates gnupg
RUN python3 -m easy_install pip
RUN pip3 install --upgrade pip
RUN pip3 install awscli --upgrade
RUN curl -L https://dl.min.io/client/mc/release/linux-amd64/mc >/usr/local/bin/mc \
  && chmod ugo+x /usr/local/bin/mc
RUN mkdir -p $HOME/.aws \
  && printf "[default]\naws_access_key_id = \naws_secret_access_key = " >$HOME/.aws/credentials


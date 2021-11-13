FROM ubuntu:20.04
LABEL maintainer=msteffen@pachyderm.io
# Fix timezone issue
ENV TZ=UTC
RUN ln -snf /usr/share/zoneinfo/$TZ /etc/localtime && echo $TZ > /etc/timezone

RUN apt-get update -y && apt-get install -y \
  curl \
  python3 \
  python3-dev \
  python3-setuptools \
  groff \
  apt-transport-https \
  ca-certificates gnupg \
  wget
RUN python3 -m easy_install pip
RUN pip3 install --upgrade pip
RUN pip3 install awscli --upgrade
RUN wget -nv -O /usr/local/bin/mc https://dl.min.io/client/mc/release/linux-amd64/mc \
  && chmod ugo+x /usr/local/bin/mc
RUN mkdir -p $HOME/.aws \
  && printf "[default]\naws_access_key_id = \naws_secret_access_key = " >$HOME/.aws/credentials


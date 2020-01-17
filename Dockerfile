FROM ubuntu:18.04
LABEL maintainer="jdoliner@pachyderm.io"

RUN \
  apt-get update -yq && \
  apt-get install -yq --no-install-recommends \
    build-essential \
    ca-certificates \
    curl \
    git \
    libssl-dev \
    pkg-config && \
  apt-get clean && \
  rm -rf /var/lib/apt
COPY etc/compile/GO_VERSION GO_VERSION
RUN \
  curl -fsSL https://get.docker.com/builds/Linux/x86_64/docker-1.12.1.tgz | tar -C /bin -xz docker/docker --strip-components=1 && \
  chmod +x /bin/docker
RUN \
  curl -sSL https://storage.googleapis.com/golang/$(cat GO_VERSION).linux-amd64.tar.gz | tar -C /tmp -xz && \
  mkdir -p /usr/local/go && \
  mv /tmp/go/bin /usr/local/go && \
  mv /tmp/go/src /usr/local/go && \
  mkdir -p /usr/local/go/pkg/tool/linux_amd64 && \
  mv /tmp/go/pkg/include /usr/local/go/pkg && \
  mv /tmp/go/pkg/tool/linux_amd64/compile /usr/local/go/pkg/tool/linux_amd64 && \
  mv /tmp/go/pkg/tool/linux_amd64/asm /usr/local/go/pkg/tool/linux_amd64 && \
  mv /tmp/go/pkg/tool/linux_amd64/link /usr/local/go/pkg/tool/linux_amd64 && \
  rm -rf /tmp/go && \
  mkdir -p /go/bin
ENV PATH /go/bin:/usr/local/go/bin:$PATH
ENV GOPATH /go
RUN go get github.com/kisielk/errcheck golang.org/x/lint/golint

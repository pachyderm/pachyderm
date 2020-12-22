FROM ubuntu:18.04
LABEL maintainer="jdoliner@pachyderm.io"

# Install FUSE
RUN \
  apt-get update -yq && \
  apt-get install -yq --no-install-recommends \
    git \
    ca-certificates \
    curl \
    fuse && \
  apt-get clean && \
  rm -rf /var/lib/apt

# Install Go 1.15.4
RUN \
  curl -sSL https://golang.org/dl/go1.15.4.linux-amd64.tar.gz | tar -C /usr/local -xz && \
  mkdir -p /go/bin
ENV PATH /usr/local/go/bin:/usr/local/bin:/usr/local/sbin:/usr/bin:/usr/sbin:/bin:/sbin
ENV GOPATH /go
ENV GOROOT /usr/local/go

# Install Pachyderm job-shim
RUN go get github.com/pachyderm/pachyderm && \
	go get github.com/pachyderm/pachyderm/src/server/cmd/job-shim && \
    cp $GOPATH/bin/job-shim /job-shim

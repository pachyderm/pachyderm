#!/bin/sh

# this script is expected to be run as root
# ${1}: user to install for

GO_VERSION=1.5.1
DOCKER_COMPOSE_VERSION=1.5.0rc2

apt-get update -yq && \
apt-get upgrade -yq && \
apt-get install -yq --no-install-recommends \
  btrfs-tools \
  build-essential \
  ca-certificates \
  cmake \
  curl \
  fuse \
  git \
  libssl-dev \
  mercurial  \
  pkg-config

# installs go
curl -sSL https://storage.googleapis.com/golang/go${GO_VERSION}.linux-amd64.tar.gz | tar -C /usr/local -xz
echo 'export PATH=${PATH}:/usr/local/go/bin' >> '/etc/profile'
su - ${1} -c "echo mkdir -p /home/${1}/go >> /home/${1}/.profile"
su - ${1} -c "echo export GOPATH=/home/${1}/go >> /home/${1}/.profile"
su - ${1} -c "echo export GO15VENDOREXPERIMENT=1 >> /home/${1}/.profile"

# installs docker
curl -sSL https://experimental.docker.com | sh
# sudoless use of docker
groupadd docker || true
usermod -aG docker ${1}
service docker restart

# installs docker-compose
curl -sSL https://github.com/docker/compose/releases/download/${DOCKER_COMPOSE_VERSION}/docker-compose-$(uname -s)-$(uname -m) > /usr/local/bin/docker-compose
chmod +x /usr/local/bin/docker-compose

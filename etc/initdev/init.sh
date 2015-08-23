#!/bin/sh

# this script is expected to be run as root
# ${1}: user to install for

GO_VERSION=1.5rc1
DOCKER_COMPOSE_VERSION=1.4.0rc3
KUBERNETES_VERSION=1.0.1

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
# we use the beta for now since it will be released in a few weeks
curl -sSL https://storage.googleapis.com/golang/go${GO_VERSION}.linux-amd64.tar.gz | tar -C /usr/local -xz
echo 'export PATH=${PATH}:/usr/local/go/bin:/home/${1}/go/bin' >> '/etc/profile'
su - ${1} -c "echo mkdir -p /home/${1}/go >> /home/${1}/.profile"
su - ${1} -c "echo export GOPATH=/home/${1}/go >> /home/${1}/.profile"

# installs docker
curl -sSL https://get.docker.com | sh
# sudoless use of docker
groupadd docker || true
usermod -aG docker ${1}
service docker restart

# installs docker-compose
curl -sSL https://github.com/docker/compose/releases/download/${DOCKER_COMPOSE_VERSION}/docker-compose-$(uname -s)-$(uname -m) > /usr/local/bin/docker-compose
chmod +x /usr/local/bin/docker-compose
# installs kubernetes cli
curl -sSL https://storage.googleapis.com/kubernetes-release/release/v${KUBERNETES_VERSION}/bin/linux/amd64/kubectl > /usr/local/bin/kubectl
chmod +x /usr/local/bin/kubectl

# installs git2go statically linked to libgit2
curl -sSL https://raw.githubusercontent.com/pachyderm/pachyderm/master/etc/git2go/install.sh | sh

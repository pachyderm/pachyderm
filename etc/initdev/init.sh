#!/bin/sh

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
  pkg-config \
  wget

# installs go
# we use the beta for now since it will be released in a few weeks
curl -sSL https://storage.googleapis.com/golang/go1.5beta3.linux-amd64.tar.gz | tar -C /usr/local -xz
echo 'export PATH=${PATH}:/usr/local/go/bin' >> '/etc/profile'
echo 'export GOROOT=/usr/local/go' >> '/etc/profile'
# makes ~/go and sets GOPATH to ~/go
su - ${1} -c "echo mkdir -p /home/${1}/go >> /home/${1}/.bash_aliases"
su - ${1} -c "echo export GOPATH=/home/${1}/go >> /home/${1}/.bash_aliases"

wget -qO- https://get.docker.com/ | sh
# sudoless use of docker
groupadd docker || true
usermod -aG docker ${1}
service docker restart

# installs docker-compose
curl -sSL https://github.com/docker/compose/releases/download/1.4.0rc2/docker-compose-$(uname -s)-$(uname -m) > /usr/local/bin/docker-compose
chmod +x /usr/local/bin/docker-compose
# installs kubernetes cli
curl -sSL https://storage.googleapis.com/kubernetes-release/release/v1.0.1/bin/linux/amd64/kubectl > /usr/local/bin/kubectl
chmod +x /usr/local/bin/kubectl

# installs git2go statically linked to libgit2
curl -sSL https://raw.githubusercontent.com/pachyderm/pachyderm/master/etc/git2go/install.sh | sh

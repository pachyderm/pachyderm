#!/bin/sh

set -e

apt-get update -y
apt-get upgrade -y
apt-get install -y btrfs-tools

go_tmpfile="/tmp/go.$$"
trap "rm -rf '${go_tmpfile}'" EXIT
curl -L "https://storage.googleapis.com/golang/go1.4.2.linux-amd64.tar.gz" > "${go_tmpfile}"
tar -C "/usr/local" -xzf "${go_tmpfile}"
echo 'export PATH=\${PATH}:/usr/local/go/bin' >> '/etc/profile'
echo 'export GOROOT=/usr/local/go' >> '/etc/profile'
su - vagrant -c "echo mkdir -p /home/vagrant/go >> /home/vagrant/.bash_aliases"
su - vagrant -c "echo export GOPATH=/home/vagrant/go >> /home/vagrant/.bash_aliases"

wget -qO- https://get.docker.com/ | sh
usermod -aG docker vagrant
service docker restart

su - vagrant -c "go get github.com/pachyderm/pachyderm"
su - vagrant -c "make -C /home/vagrant/go/src/github.com/pachyderm/pachyderm test-deps"

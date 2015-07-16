#!/bin/sh

set -e

GO_VERSION="1.4.2"

apt-get update -y
apt-get upgrade -y
apt-get install -y btrfs-tools

go_tmpfile="/tmp/go.$$"
trap "rm -rf '${go_tmpfile}'" EXIT
curl -L "https://storage.googleapis.com/golang/go${GO_VERSION}.linux-amd64.tar.gz" > "${go_tmpfile}"
tar -C "/usr/local" -xzf "${go_tmpfile}"
#su - root -c "echo 'export PATH=\${PATH}:/usr/local/go/bin' >> '/etc/profile'"
#su - root -c "echo 'export GOROOT=/usr/local/go' >> '/etc/profile'"
echo 'export PATH=\${PATH}:/usr/local/go/bin' >> '/etc/profile'
echo 'export GOROOT=/usr/local/go' >> '/etc/profile'
su - vagrant -c "echo mkdir -p /home/vagrant/go >> /home/vagrant/.bash_aliases"
su - vagrant -c "echo export GOPATH=/home/vagrant/go >> /home/vagrant/.bash_aliases"

wget -qO- https://get.docker.com/ | sudo sh
#groupadd docker || true
#su - vagrant -c usermod -aG docker "${USER}" || true
#service docker restart

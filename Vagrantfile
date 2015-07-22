# -*- mode: ruby -*-
# vi: set ft=ruby :

VAGRANTFILE_API_VERSION = "2"

INIT_SCRIPT = <<SCRIPT
apt-get update -yq && \
apt-get upgrade -yq && \
apt-get install -yq --no-install-recommends \
  btrfs-tools \
  build-essential \
  ca-certificates \
  curl \
  libgit2-dev \
  pkg-config \
  git

curl -sSL https://storage.googleapis.com/golang/go1.4.2.linux-amd64.tar.gz | tar -C /usr/local -xz
mkdir -p /go/bin
echo 'export PATH=${PATH}:/usr/local/go/bin' >> '/etc/profile'
echo 'export GOROOT=/usr/local/go' >> '/etc/profile'
su - vagrant -c "echo mkdir -p /home/vagrant/go >> /home/vagrant/.bash_aliases"
su - vagrant -c "echo export GOPATH=/home/vagrant/go >> /home/vagrant/.bash_aliases"

wget -qO- https://get.docker.com/ | sh
usermod -aG docker vagrant
service docker restart
SCRIPT

Vagrant.configure(VAGRANTFILE_API_VERSION) do |config|
  config.vm.box = "ubuntu/vivid64"

  config.vm.provider "virtualbox" do |vb|
    vb.customize ["modifyvm", :id, "--memory", "2048"]
    vb.customize ["modifyvm", :id, "--cpus", "2"]
    vb.customize ["modifyvm", :id, "--natdnshostresolver1", "on"]
    vb.customize ["modifyvm", :id, "--natdnsproxy1", "on"]
  end
  config.vm.synced_folder ENV['HOME'], "/home/vagrant/host"

  config.vm.provision "shell", inline: INIT_SCRIPT
end

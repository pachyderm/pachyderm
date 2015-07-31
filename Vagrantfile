# -*- mode: ruby -*-
# vi: set ft=ruby :

VAGRANTFILE_API_VERSION = "2"

INIT_SCRIPT = <<SCRIPT
apt-get update -yq && \
apt-get upgrade -yq && \
apt-get install -yq --no-install-recommends \
  btrfs-tools \
  build-essential \
  curl \
  libgit2-dev \
  pkg-config \
  git \
  fuse

# installs go
# we use the beta for now since it will be released in a few weeks
curl -sSL https://storage.googleapis.com/golang/go1.5beta3.linux-amd64.tar.gz | tar -C /usr/local -xz
mkdir -p /go/bin
echo 'export PATH=${PATH}:/usr/local/go/bin' >> '/etc/profile'
echo 'export GOROOT=/usr/local/go' >> '/etc/profile'
# makes ~/go and sets GOPATH to ~/GO
su - vagrant -c "echo mkdir -p /home/vagrant/go >> /home/vagrant/.bash_aliases"
su - vagrant -c "echo export GOPATH=/home/vagrant/go >> /home/vagrant/.bash_aliases"

wget -qO- https://get.docker.com/ | sh
# sudoless use of docker
groupadd docker || true
usermod -aG docker vagrant
service docker restart

# installs docker-compose
curl -sSL https://github.com/docker/compose/releases/download/1.4.0rc2/docker-compose-$(uname -s)-$(uname -m) > /usr/local/bin/docker-compose
chmod +x /usr/local/bin/docker-compose
# installs kubernetes cli
curl -sSL https://storage.googleapis.com/kubernetes-release/release/v1.0.1/bin/linux/amd64/kubectl > /usr/local/bin/kubectl
chmod +x /usr/local/bin/kubectl
SCRIPT

Vagrant.configure(VAGRANTFILE_API_VERSION) do |config|
  config.vm.box = "ubuntu/vivid64" # 15.04

  config.vm.provider "virtualbox" do |vb|
    vb.customize ["modifyvm", :id, "--memory", "2048"]
    vb.customize ["modifyvm", :id, "--cpus", "2"]
    # https://stefanwrobel.com/how-to-make-vagrant-performance-not-suck
    vb.customize ["modifyvm", :id, "--natdnshostresolver1", "on"]
    vb.customize ["modifyvm", :id, "--natdnsproxy1", "on"]
  end

  # allows calling docker from host machine if networking properly set up
  # see https://github.com/peter-edge/dotfiles/tree/master/bin/setup_docker_http.sh
  # note that the above script should not be used if your vagrant box is somehow insecure
  # see https://github.com/peter-edge/dotfiles/tree/master/bash_aliases_docker up until
  # the 'alias docker' command for how to set this all up with a mac (after 'brew install docker')
  config.vm.network "private_network", ip: "192.168.10.10"
  # allows volumes from host machine if calling docker from host machine
  config.vm.synced_folder ENV['HOME'], ENV['HOME']

  config.vm.provision "shell", inline: INIT_SCRIPT
end

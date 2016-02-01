#!/bin/bash

# this script is expected to be run as root
# ${1}: user to install for

set -o nounset		# using an unset variable is an error
set -o errexit		# exit with a nonzero exit status on any unhandled error

USERNAME="${1}"
GO_VERSION=1.5.1
DOCKER_COMPOSE_VERSION=1.5.0rc2

# to determine which Ubuntu version we're running on for updating the apt sources
. /etc/lsb-release

setup_apt()
{
	# Adds the correct docker apt repository
	# From instructions at https://docs.docker.com/engine/installation/ubuntulinux/
	apt-key adv --keyserver hkp://p80.pool.sks-keyservers.net:80 --recv-keys 58118E89F3A912897C070ADBF76221572C52609D
	rm -f /etc/apt/sources.list.d/docker.list
	echo "deb https://apt.dockerproject.org/repo ubuntu-${DISTRIB_CODENAME} main" > /etc/apt/sources.list.d/docker.list
}

install_docker_compose()
{
	local URL="https://github.com/docker/compose/releases/download/${DOCKER_COMPOSE_VERSION}/docker-compose-$(uname -s)-$(uname -m)"

	curl -sSL "${URL}" > /usr/local/bin/docker-compose
	chmod +x /usr/local/bin/docker-compose
}

install_go()
{
	# installs go
	curl -sSL https://storage.googleapis.com/golang/go${GO_VERSION}.linux-amd64.tar.gz | tar -C /usr/local -xz

	# configures vagrant user's environment for go development
	echo 'export PATH=${PATH}:/usr/local/go/bin' >> '/etc/profile'
	su - "${USERNAME}" -c "echo mkdir -p /home/${USERNAME}/go >> /home/${USERNAME}/.profile"
	su - "${USERNAME}" -c "echo export GOPATH=/home/${USERNAME}/go >> /home/${USERNAME}/.profile"
	su - "${USERNAME}" -c "echo export PATH=/home/${USERNAME}/go/bin:"'${PATH}'" >> /home/${USERNAME}/.profile"
	su - "${USERNAME}" -c "echo export GO15VENDOREXPERIMENT=1 >> /home/${USERNAME}/.profile"
}

install_apt()
{
	apt-get update -yq
	apt-get upgrade -yq
	apt-get install -yq --no-install-recommends \
		build-essential \
		ca-certificates \
		cmake \
		curl \
		fuse \
		git \
		libssl-dev \
		mercurial  \
		pkg-config \
		docker-engine

}

cleanup_apt()
{
	apt-get autoremove -yq
}

run()
{
	echo "Running $*" 1>&2
	$*
}

main()
{
	run setup_apt
	run install_apt
	run install_go
	run install_docker_compose
	run cleanup_apt
}

# # sudoless use of docker
# groupadd docker || true
# usermod -aG docker ${1}
# service docker restart

main

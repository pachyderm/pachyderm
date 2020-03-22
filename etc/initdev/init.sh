#!/bin/bash

# this script is expected to be run as root
# ${1}: user to install for

set -o nounset		# using an unset variable is an error
set -o errexit		# exit with a nonzero exit status on any unhandled error

USERNAME="${1}"
GO_VERSION=1.7.3

# to determine which Ubuntu version we're running on for updating the apt sources
# shellcheck disable=SC1091
. /etc/lsb-release

run()
{
	echo "Running $*" 1>&2
	# shellcheck disable=SC2048
	$*
}

setup_apt()
{
	# Adds the correct docker apt repository
	# From instructions at https://docs.docker.com/engine/installation/ubuntulinux/
	apt-key adv --keyserver hkp://p80.pool.sks-keyservers.net:80 --recv-keys 58118E89F3A912897C070ADBF76221572C52609D
	rm -f /etc/apt/sources.list.d/docker.list
	echo "deb https://apt.dockerproject.org/repo ubuntu-${DISTRIB_CODENAME} main" > /etc/apt/sources.list.d/docker.list
}

install_go()
{
	curl -sSL https://storage.googleapis.com/golang/go${GO_VERSION}.linux-amd64.tar.gz | tar -C /usr/local -xz
}

# configures vagrant user's environment for go development
setup_user_go()
{
	# shellcheck disable=SC2016
	echo 'export PATH=${PATH}:/usr/local/go/bin' >> '/etc/profile'
	su - "${USERNAME}" -c "echo mkdir -p /home/${USERNAME}/go >> /home/${USERNAME}/.profile"
	su - "${USERNAME}" -c "echo export GOPATH=/home/${USERNAME}/go >> /home/${USERNAME}/.profile"
	# shellcheck disable=SC2016
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

# sudoless use of docker
setup_user_docker()
{
	groupadd docker || true
	usermod -aG docker "${USERNAME}"
	service docker restart
}

setup_user()
{
	run setup_user_docker
	run setup_user_go
}

main()
{
	run setup_apt

	run install_apt
	run install_go

	run cleanup_apt

	run setup_user
}


main

#!/bin/bash

# Check if this script has already been run

which nvidia-smi
if [[ $? -eq 0 ]]; then
	# If it has, do nothing
	echo "Nvidia drivers already installed. Nothing to do"
else
	set +euxo pipefail
	# If not ... install the drivers and restart
	NVIDIA_RUNNER=$1
	apt-get update
	apt-get install --yes gcc
    ./$NVIDIA_RUNNER --ui=none --no-questions --accept-license
	# We want to overwrite, not append since the default
	# value of this file w our current AMI is just 'exit 0'
	echo "PACHNVIDIADRIVERINSTALL updating rc.local"
    cp /etc/rc.local /etc/rc.local.bck
	cat >/etc/rc.local  <<EOL
nvidia-smi -pm 1 || true
nvidia-smi -acp 0 || true
nvidia-smi --auto-boost-default=0 || true
nvidia-smi --auto-boost-permission=0 || true
nvidia-modprobe -u -c=0 -m || true
EOL
	echo "Updated /etc/rc.local:"
	cat /etc/rc.local

	echo "PACHNVIDIADRIVERINSTALL (not) going to restart"
	# Don't think this will work ... but it might
	# if not ... our deploy script / instructions will need to include doing a restart
	# only AFTER the driver install has completed
	/sbin/shutdown -r	
fi

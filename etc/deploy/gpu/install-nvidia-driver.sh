#!/bin/bash

# Check if this script has already been run
chroot /rootfs which nvidia-smi
if [[ $? -eq 0 ]]; then
	# If it has, do nothing
	echo "Nvidia drivers already installed. Nothing to do"
else
	set +euxo pipefail
	# If not ... install the drivers and restart
	echo $1
	NVIDIA_RUNNER=$1
	echo "user id:"
	id -u
	echo "whoami:"
	whoami
	echo "write a single thing"
	chroot /rootfs echo "hallo" > /tmp/ohhai
	echo "exit code: $?"
    echo "wrote:"
    cat /rootfs/tmp/ohhai
    pwd
    cat /tmp/ohhai
	chroot /rootfs echo "hallo" > /ohhai
	echo "exit code: $?"
    echo "wrote:"
    cat /rootfs/ohhai
    pwd
    cat /ohhai
	echo "write without chroot"
	echo "third times the charm" > /rootfs/tmp/ohhaiz
	echo "exit code: $?"
	echo "file contents of /tmp/ohhai"
	cat /tmp/ohhai
    echo "file contents of /rootfs/tmp/ohhaiz"
    cat /rootfs/tmp/ohhaiz
	cp $NVIDIA_RUNNER /rootfs
	chroot /rootfs apt-get update
	cp /etc/sudoers /rootfs/etc/sudoers
	chroot /rootfs apt-get install --yes gcc
# disable the actual install while I debug the rc local bit
#	chroot /rootfs ./$NVIDIA_RUNNER --ui=none --no-questions --accept-license
	# We want to overwrite, not append since the default
	# value of this file w our current AMI is just 'exit 0'
	echo "PACHNVIDIADRIVERINSTALL updating rc.local"
	chroot /rootfs sudo cat >>/etc/rc.local  <<EOL
nvidia-smi -pm 1 || true
nvidia-smi -acp 0 || true
nvidia-smi --auto-boost-default=0 || true
nvidia-smi --auto-boost-permission=0 || true
nvidia-modprobe -u -c=0 -m || true
EOL
	echo "state of /etc/rc.local:"
	chroot /rootfs cat /etc/rc.local
	echo 'trying without sudo'
	chroot /rootfs cat >>/etc/rc.local  <<EOL
nvidia-smi -pm 1 || true
nvidia-smi -acp 0 || true
nvidia-smi --auto-boost-default=0 || true
nvidia-smi --auto-boost-permission=0 || true
nvidia-modprobe -u -c=0 -m || true
EOL
	echo "state of /etc/rc.local:"
	chroot /rootfs cat /etc/rc.local
	echo 'trying to write to tmp'
	chroot /rootfs cat >>/tmp/rc.local  <<EOL
nvidia-smi -pm 1 || true
nvidia-smi -acp 0 || true
nvidia-smi --auto-boost-default=0 || true
nvidia-smi --auto-boost-permission=0 || true
nvidia-modprobe -u -c=0 -m || true
EOL
	echo "state of /etc/rc.local:"
	chroot /rootfs cat /tmp/rc.local

	echo 'trying with sudo at the start'
	sudo chroot /rootfs cat >>/etc/rc.local  <<EOL
nvidia-smi -pm 1 || true
nvidia-smi -acp 0 || true
nvidia-smi --auto-boost-default=0 || true
nvidia-smi --auto-boost-permission=0 || true
nvidia-modprobe -u -c=0 -m || true
EOL
	echo "state of /etc/rc.local:"
	chroot /rootfs cat /etc/rc.local

	echo "PACHNVIDIADRIVERINSTALL (not) going to restart"
	# Don't think this will work ... but it might
	# if not ... our deploy script / instructions will need to include doing a restart
	# only AFTER the driver install has completed
	#chroot /rootfs shutdown -r	
fi

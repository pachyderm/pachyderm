#!/bin/sh

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
	chroot /rootfs ./$NVIDIA_RUNNER --ui=none --no-questions --accept-license
	# We want to overwrite, not append since the default
	# value of this file w our current AMI is just 'exit 0'
	chroot /rootfs sudo cat >>/etc/rc.local  <<EOL
nvidia-smi -pm 1 || true
nvidia-smi -acp 0 || true
nvidia-smi --auto-boost-default=0 || true
nvidia-smi --auto-boost-permission=0 || true
nvidia-modprobe -u -c=0 -m || true
EOL

	# Don't think this will work ... but it might
	# if not ... our deploy script / instructions will need to include doing a restart
	# only AFTER the driver install has completed
	chroot /rootfs shutdown -r	
fi

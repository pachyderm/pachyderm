#!/bin/bash

# Check if this script has already been run
which nvidia-smi
if [[ $? -eq 0 ]]; then
	# If it has, do nothing
	echo "Drivers already installed. Nothing to do"
else
	set +euxo pipefail
	# If not ... install the drivers and restart
	echo $NVIDIA_RUNNER
	./$NVIDIA_RUNNER --ui=none --no-questions --accept-license
	# We want to overwrite, not append since the default
	# value of this file w our current AMI is just 'exit 0'
	sudo cat >>/etc/rc.local  <<EOL
nvidia-smi -pm 1 || true
nvidia-smi -acp 0 || true
nvidia-smi --auto-boost-default=0 || true
nvidia-smi --auto-boost-permission=0 || true
nvidia-modprobe -u -c=0 -m || true
EOL

	shutdown -r	
fi

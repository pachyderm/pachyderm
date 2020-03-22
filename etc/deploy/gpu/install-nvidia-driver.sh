#!/bin/bash

# Check if this script has already been run
prefix="Pachyderm nvidia bootstrapper"
command -v nvidia-smi
if [[ $? -eq 0 ]]; then
    # If it has, do nothing
    echo "$prefix : Nvidia drivers already installed. Nothing to do"
    exit 0
else
    apt-get install pciutils -y
    lspci -vnn | grep NVIDIA
    if [[ $? -ne 0 ]]; then
        # No NVIDIA card present, exit
        echo "$prefix: No NVIDIA GPU detected, exiting"
        exit 0
    fi
    set +euxo pipefail
    # If not ... install the drivers and restart
    NVIDIA_RUNNER=$1
    apt-get update
    apt-get install --yes gcc
    "./$NVIDIA_RUNNER" --ui=none --no-questions --accept-license
    # We want to overwrite, not append since the default
    # value of this file w our current AMI is just 'exit 0'
    echo "$prefix : Nvidia driver bootstrapper - updating rc.local"
    cp /etc/rc.local /etc/rc.local.bck
    cat >/etc/rc.local  <<EOL
#!/bin/sh -e
nvidia-smi -pm 1 || true
nvidia-smi -acp 0 || true
nvidia-smi --auto-boost-default=0 || true
nvidia-smi --auto-boost-permission=0 || true
nvidia-modprobe -u -c=0 -m || true
EOL
    echo "$prefix : Updated /etc/rc.local:"
    cat /etc/rc.local
    /sbin/shutdown -r   
fi

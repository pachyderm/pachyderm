#!/bin/sh

# If the guest container does not have SSL certificates, use our own
# certificates
if [ ! -d "/etc/ssl/certs" ]; then
    mkdir -p /etc/ssl
    cp -r /pach-bin/certs /etc/ssl/certs
fi

/pach-bin/worker $1

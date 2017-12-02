#!/bin/bash

set -Eex

echo 'looking for resolv.conf files'
echo 'run/systemd/resolve/resolv.conf:'
cat /run/systemd/resolve/resolv.conf || true
echo '/etc/systemd/resolved.conf:'
cat /etc/systemd/resolved.conf || true

touch /etc/resolve.conf
echo 'pre resolv conf:'
cat /etc/resolve.conf || true
echo 'nameserver 8.8.8.8' > /etc/resolv.conf
echo 'post resolv conf:'
cat /etc/resolve.conf || true
sudo systemctl stop systemd-resolved || true
sudo systemctl disable systemd-resolved || true

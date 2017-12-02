#!/bin/bash


touch /etc/resolve.conf
echo 'pre resolv conf:'
cat /etc/resolve.conf || true
echo 'nameserver 8.8.8.8' > /etc/resolv.conf
echo 'post resolv conf:'
cat /etc/resolve.conf || true
systemctl stop systemd-resolved
systemctl disable systemd-resolved

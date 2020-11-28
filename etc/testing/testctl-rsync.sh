#!/bin/bash
set -euo pipefail
IP=$(grep "##EXTERNAL_IP=" "$KUBECONFIG" |cut -d '=' -f 2-)
PORT=$(grep "##SSH_FORWARDED_PORT=" "$KUBECONFIG" |cut -d '=' -f 2-)
grep "##PRIVATE_KEY_ONELINER=" "$KUBECONFIG" |cut -d '=' -f 2- |base64 -d > id_rsa
chmod 0600 id_rsa
local_path="$1"
remote_path="$2"
exec rsync -a --delete -e "ssh -i $(pwd)/id_rsa -p $PORT -o UserKnownHostsFile=/dev/null -o StrictHostKeyChecking=no" "$local_path" root@"$IP":"$remote_path"

#!/bin/bash
echo "{\"pachd_address\": \"$PACHD_CLUSTER_ADDRESS\"}" | pachctl config set context in-cluster --overwrite
pachctl config set active-context in-cluster
echo "$1" | pachctl auth login -t
git clone https://github.com/pachyderm/examples
sudo chown -R jovyan examples
sudo chown -R jovyan /home/jovyan/.pachyderm

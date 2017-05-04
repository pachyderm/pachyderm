#!/bin/bash

#if [[ "$TRAVIS_EVENT_TYPE" != "cron" ]]; then
if [[ "cron" != "cron" ]]; then
  exit 0
fi

# Install kops
wget https://github.com/kubernetes/kops/releases/download/1.5.1/kops-linux-amd64
chmod +x kops-linux-amd64
sudo mv kops-linux-amd64 /usr/local/bin/kops

# Install aws (needed by batch test suite)
python3 --version
pip --version
pip install awscli --upgrade --user
aws --version

# Install gcloud and gsutil
wget https://dl.google.com/dl/cloudsdk/channels/rapid/downloads/google-cloud-sdk-153.0.0-linux-x86_64.tar.gz \
    -O google-cloud-sdk-153.0.0-linux-x86_64.tar.gz
tar -xzf google-cloud-sdk-153.0.0-linux-x86_64.tar.gz
mv google-cloud-sdk/bin/gcloud /usr/local/bin/gcloud
mv google-cloud-sdk/bin/gsutil /usr/local/bin/gsutil

# because we need to run the batch test suite with sudo, we need to make sure
# a few binaries are in the secure path
sudo bash -Ec 'echo $PATH'
sudo ln -s `which aws` /usr/local/bin/aws
sudo ln -s `which go` /usr/local/bin/go
sudo apt-get install -yq uuid

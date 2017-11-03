#!/bin/bash

set -e

sudo apt-get install silversearcher-ag
echo "Running local tests"
make test

# Disable aws CI for now, see:
# https://github.com/pachyderm/pachyderm/issues/2109
exit 0

echo "Running local tests"
make local-test

echo "Running aws tests"

sudo pip install awscli
sudo apt-get install realpath uuid jq
wget --quiet https://github.com/kubernetes/kops/releases/download/1.6.2/kops-linux-amd64
chmod +x kops-linux-amd64
sudo mv kops-linux-amd64 /usr/local/bin/kops

# Use the secrets in the travis environment to setup the aws creds for the aws command:
echo -e "${AWS_ACCESS_KEY_ID}\n${AWS_SECRET_ACCESS_KEY}\n\n\n" \
    | aws configure

make install
echo "pachctl is installed here:"
which pachctl

# Travis doesn't come w an ssh key
# kops needs one in place (because it enables ssh access to nodes w it)
# for now we'll just generate one on the fly
# travis supports adding a persistent one if we pay: https://docs.travis-ci.com/user/private-dependencies/#Generating-a-new-key
if [[ ! -e ${HOME}/.ssh/id_rsa ]]; then
    ssh-keygen -t rsa -b 4096 -C "buildbot@pachyderm.io" -f ${HOME}/.ssh/id_rsa -N ''
    echo "generated ssh keys:"
    ls ~/.ssh
    cat ~/.ssh/id_rsa.pub
fi

# Need to login so that travis can push the bench image
docker login -u pachydermbuildbot -p ${DOCKER_PWD}

# Run tests in the cloud
make aws-test

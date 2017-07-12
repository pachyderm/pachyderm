#!/bin/bash

echo "Running local tests"
make test

echo "Running aws tests"

pip install awscli

# Use the secrets in the travis environment to setup the aws creds for the aws command:
echo -e "${AWS_ACCESS_KEY_ID}\n${AWS_SECRET_ACCESS_KEY}\n" \
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
sudo -E PATH="${PATH}" GOPATH="${GOPATH}" make aws-test

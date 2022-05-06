#!/bin/bash

set -ex

sudo apt-get update
sudo apt-get -y install python-virtualenv

make ./integration/python/venv

pushd ~
    mkdir -p bin
    pushd bin
        wget https://dl.min.io/client/mc/release/linux-amd64/archive/mc.RELEASE.2020-05-06T18-00-07Z
        mv mc.RELEASE.2020-05-06T18-00-07Z mc
        chmod +x mc
    popd

    curl "https://awscli.amazonaws.com/awscli-exe-linux-x86_64.zip" -o "awscliv2.zip"
    unzip awscliv2.zip
    sudo ./aws/install
popd

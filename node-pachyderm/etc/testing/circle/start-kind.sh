#!/bin/bash

set -ex

export PATH=$(pwd):$(pwd)/cached-deps:$GOPATH/bin:$PATH

kind create cluster && kubectl config set current-context kind-kind

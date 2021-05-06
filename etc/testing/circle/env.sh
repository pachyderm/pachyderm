#!/bin/bash

GOCACHE=/home/circleci/.gocache
export GOCACHE

GOPATH=/home/circleci/.go_workspace
export GOPATH

PATH=$(pwd):$(pwd)/cached-deps:$GOPATH/bin:$PATH
export PATH


#!/bin/bash

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"
cd "$DIR"/../..

GOPATH=/root/go
export GOPATH
PATH="${GOPATH}/bin:${PATH}"
export PATH

go get -u github.com/jstemmer/go-junit-report
go-junit-report < test.out 

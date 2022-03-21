#!/bin/bash
set -xeuo pipefail

while true; do

    pachctl create repo data
    sleep 1
    echo "version 1" | pachctl put file data@master:/myfile.txt
    sleep 1
    curl -XPUT 'localhost:9002/repos/data/master/_mount?name=data&mode=ro'
    sleep 1
    cat /pfs/data2/myfile.txt
    sleep 1
    curl -XPUT 'localhost:9002/repos/data/master/_unmount?name=data'
    sleep 1
    pachctl create branch data@v1 --head master
    sleep 1
    curl -XPUT 'localhost:9002/repos/data/master/_mount?name=data&mode=ro'
    sleep 1
    cat /pfs/data2/myfile.txt
    sleep 1
    pachctl delete repo data
    sleep 1

done

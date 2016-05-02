#!/bin/bash

umount $HOME/pfs
rm -rf $HOME/pfs
make clean-launch
docker-machine restart dev
eval "$(docker-machine env dev)"
docker-machine regenerate-certs dev
eval "$(docker-machine env dev)"
ps -ef | grep "docker/machine" | grep -v grep | cut -f 4 -d " " | while read -r pid
do
    kill $pid
done
docker-machine ssh dev -fTNL 8080:localhost:8080 -L 30650:localhost:30650
docker ps

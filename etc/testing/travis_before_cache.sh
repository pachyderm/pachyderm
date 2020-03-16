#!/bin/bash
# Caches docker images. Note that, at the moment, only `pachyderm_build`
# is cached. That's because the pachd/worker images are going to be
# invalidated with any code changes, and saving/restoring those images will
# just slow things down.
mkdir -p $HOME/docker
rm $HOME/docker/*
tag=pachyderm_build:latest
id=$(docker images pachyderm_build --format '{{.ID}}')
test -e $HOME/docker/${id}.tar.gz || docker save ${tag} | gzip -2 > $HOME/docker/${id}.tar.gz

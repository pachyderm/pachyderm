#!/bin/bash
mkdir -p "$HOME/docker"
rm "$HOME/docker/*"
tag=pachyderm_build:latest
id=$(docker images pachyderm_build --format '{{.ID}}')
test -e "$HOME/docker/${id}.tar.gz" || docker save ${tag} | gzip -2 > "$HOME/docker/${id}.tar.gz"

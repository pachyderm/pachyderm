#!/bin/bash

set -xe

if [ -n "${SETUP}" ]; then
  pfs create-repo test
  commit_id="$(pfs start-commit test scratch)"
  echo hello | pfs put-file test ${commit_id} foo.txt
  pfs finish-commit test ${commit_id}
else
  commit_id="$(pfs list-commit test | head -2 | tail -1 | cut -f 1 -d ' ')"
fi
pfs list-file test ${commit_id} /
if [ -n "${DOCKER_VOLUME}" ]; then
  docker volume create --driver=pfs --opt repository=test --opt commit_id=${commit_id} --opt shard=0 --opt modulus=1 --name foo
  docker run --name foo --volume foo:/in --volume-driver pfs ubuntu cat /in/foo.txt
  docker rm foo
  docker volume rm foo
fi

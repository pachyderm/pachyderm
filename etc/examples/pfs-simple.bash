#!/bin/bash

set -xe

pfs init test
commit_id="$(pfs branch test scratch)"
echo hello | pfs put test ${commit_id} foo.txt
pfs commit test ${commit_id}
pfs ls test ${commit_id} /
docker volume create --driver=pfs --opt repository=test --opt commit_id=${commit_id} --opt shard=0 --opt modulus=1 --name foo

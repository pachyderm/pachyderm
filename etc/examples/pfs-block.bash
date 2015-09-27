#!/bin/bash

set -xe

pfs create-repo test
commit1="$(pfs start-commit test scratch)"
block1="$(echo "hello world\n" | pfs put-block test ${commit1})"
block2="$(echo "lorem ipsum\n" | pfs put-block test ${commit1})"
pfs finish-commit test ${commit1}

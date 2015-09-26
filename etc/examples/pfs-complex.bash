#!/bin/bash

set -xe

pfs create-repo test
commit1="$(pfs start-commit test scratch)"
echo "hello world\n" | pfs put-file test ${commit1} hello.txt
echo "lorem ipsum\n" | pfs put-file test ${commit1} lorem.txt
pfs finish-commit test ${commit1}
commit2="$(pfs start-commit test $commit1)"
commit3="$(pfs start-commit test $commit1)"
echo "hello world\nthis is pfs\n" | pfs put-file test ${commit2} hello.txt
echo "lorem ipsum\nNeque porro\n" | pfs put-file test ${commit3} lorem.txt
pfs finish-commit test ${commit2}
pfs finish-commit test ${commit3}
commit4="$(pfs start-commit test $commit3)"
echo "fizz buzz\nfizz buzz\n" | pfs put-file test ${commit4} fizz_buzz.txt
echo "foo bar\nfoo bar\n" | pfs put-file test ${commit4} foo_bar.txt
pfs finish-commit test ${commit4}

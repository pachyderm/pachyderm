#!/bin/bash

set -xe

pfs repo-create test
commit1="$(pfs commit-start test scratch)"
echo "hello world\n" | pfs file-put test ${commit1} hello.txt
echo "lorem ipsum\n" | pfs file-put test ${commit1} lorem.txt
pfs commit-finish test ${commit1}
commit2="$(pfs commit-start test $commit1)"
commit3="$(pfs commit-start test $commit1)"
echo "hello world\nthis is pfs\n" | pfs file-put test ${commit2} hello.txt
echo "lorem ipsum\nNeque porro\n" | pfs file-put test ${commit3} lorem.txt
pfs commit-finish test ${commit2}
pfs commit-finish test ${commit3}
commit4="$(pfs commit-start test $commit3)"
echo "fizz buzz\nfizz buzz\n" | pfs file-put test ${commit4} fizz_buzz.txt
echo "foo bar\nfoo bar\n" | pfs file-put test ${commit4} foo_bar.txt
pfs commit-finish test ${commit4}

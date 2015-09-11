#!/bin/bash

set -xe

pfs init test
commit1="$(pfs branch test scratch)"
echo "hello world\n" | pfs put test ${commit1} hello.txt
echo "lorem ipsum\n" | pfs put test ${commit1} lorem.txt
pfs commit test ${commit1}
commit2="$(pfs branch test $commit1)"
commit3="$(pfs branch test $commit1)"
echo "hello world\nthis is pfs\n" | pfs put test ${commit2} hello.txt
echo "lorem ipsum\nNeque porro\n" | pfs put test ${commit3} lorem.txt
pfs commit test ${commit2}
pfs commit test ${commit3}
commit4="$(pfs branch test $commit3)"
echo "fizz buzz\nfizz buzz\n" | pfs put test ${commit4} fizz_buzz.txt
echo "foo bar\nfoo bar\n" | pfs put test ${commit4} foo_bar.txt
pfs commit test ${commit4}

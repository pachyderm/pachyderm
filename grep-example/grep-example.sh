#!/bin/bash

echo "World"
grep -r "foo" /pfs/data > /pfs/out/file1
echo "Hello" > /pfs/out/file2

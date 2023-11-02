#!/bin/bash
# This script reads a tar of the pachyderm_sdk python package from stdin
# and prints the generated HTML files as a tarball to stdout.
#
# Note: If you are trying to develop on/debug this script, you should
# print to stderr as this will not corrupt the output.
set -e

# Extracts from stdin into ./pachyderm_sdk
tar xf /dev/stdin

# Writes HTML files to docs/pachyderm_sdk
python docs/generate.py

tar cf - docs/pachyderm_sdk

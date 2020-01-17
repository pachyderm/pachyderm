#!/bin/bash

# This script installs the python modules needed to build our docs, for local
# testing

pip3 install --user \
  mkdocs \
  mkdocs-exclude \
  mkdocs-material \
  pymdown-extensions \
  mdx-truly-sane-lists


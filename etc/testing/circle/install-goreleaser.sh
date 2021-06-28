#!/bin/bash

set -ex

mkdir -p cached-deps

# Install goreleaser 
if [ ! -f cached-deps/goreleaser ]; then
  GORELEASER_VERSION=0.169.0
  curl -L https://github.com/goreleaser/goreleaser/releases/download/v${GORELEASER_VERSION}/goreleaser_Linux_x86_64.tar.gz \
      | tar xzf - -C cached-deps goreleaser
fi

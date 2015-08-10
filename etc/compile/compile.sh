#!/bin/sh

set -Ee

export CGO_LDFLAGS="/usr/local/lib/libgit2.a $(pkg-config --libs --static /usr/local/lib/pkgconfig/libgit2.pc) \
/usr/lib/x86_64-linux-gnu/libssl.a $(pkg-config --libs --static /usr/lib/x86_64-linux-gnu/pkgconfig/libssl.pc)"

go build \
  -a \
  -installsuffix netgo \
  -tags netgo \
  -o /compile/${1} \
  src/cmd/${1}/main.go

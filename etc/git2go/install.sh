#!/bin/sh

set -Ee

INSTALL_DIR=/usr/local
VERSION=0.22.3

for executable in cmake go pkg-config uname; do
  if ! which ${executable} > /dev/null; then
    echo "error: ${executable} not installed" >&2
    exit 1
  fi
done

TMPDIR=/tmp/libgit2-install.$$
mkdir -p ${TMPDIR}
trap "rm -rf ${TMPDIR}" EXIT

curl -sSL https://codeload.github.com/libgit2/libgit2/tar.gz/v${VERSION} | tar -C ${TMPDIR} -xz

cd ${TMPDIR}/libgit2-${VERSION}
mkdir -p build
cd build

CMAKE_FLAGS="-DBUILD_CLAR=OFF \
  -DBUILD_SHARED_LIBS=OFF \
  -DCMAKE_BUILD_TYPE=Release \
  -DCMAKE_C_FLAGS=-fPIC \
  -DCMAKE_INSTALL_PREFIX=${INSTALL_DIR} \
  -DTHREADSAFE=ON"

if [ "$(uname -s)" = "Darwin" ]; then
  if ! which brew > /dev/null; then 
    echo "error: brew not installed" >&2
    exit 1
  fi
  brew install openssl
  OPENSSL_ROOT_DIR=$(brew --prefix openssl)
  CMAKE_FLAGS="${CMAKE_FLAGS} -DCMAKE_OSX_ARCHITECTURES='i386;x86_64' -DOPENSSL_ROOT_DIR=${OPENSSL_ROOT_DIR}"
  CGO_LDFLAGS="${OPENSSL_ROOT_DIR}/lib/libssl.a $(pkg-config --libs --static ${OPENSSL_ROOT_DIR}/lib/pkgconfig/libssl.pc)"
else
  # TODO(pedge)
  echo "warning: make sure you have libssl-dev installed" >&2
fi

cmake .. ${CMAKE_FLAGS}
cmake --build . --target install

export CGO_LDFLAGS="${CGO_LDFLAGS} ${INSTALL_DIR}/lib/libgit2.a $(pkg-config --libs --static ${INSTALL_DIR}/lib/pkgconfig/libgit2.pc)"
go get -d gopkg.in/libgit2/git2go.v22
go clean -i gopkg.in/libgit2/git2go.v22
go install gopkg.in/libgit2/git2go.v22

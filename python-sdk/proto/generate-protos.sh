#!/bin/bash
set -e

SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )

cd "${SCRIPT_DIR}"
docker build -t pachyderm_python_proto:python-sdk .

cd ../..
find src -regex ".*\.proto" \
  | grep -v 'internal' \
  | grep -v 'server' \
  | grep -v 'protoextensions' \
  | grep -v 'proxy' \
  | xargs tar cf - \
  | docker run -i pachyderm_python_proto:python-sdk \
  | tar -C python-sdk/pachyderm_sdk -xf -

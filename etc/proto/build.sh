#!/bin/bash
# This script build protos using the pachyderm_proto image.

if [[ "$OSTYPE" == "darwin"* ]]; then
  TAR=gtar # brew install gnu-tar if on mac
else
  TAR=tar
fi

read -ra proto_files < <(find src -name "*.proto" -print0 | xargs -0)
$TAR cf - "${proto_files[@]}" \
 | docker run -i --rm pachyderm_proto \
 | $TAR xf -

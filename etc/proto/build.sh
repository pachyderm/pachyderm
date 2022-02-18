#!/bin/bash
# This script build protos using the pachyderm_proto image.

read -ra proto_files < <(find src -name "*.proto" -print0 | xargs -0)
tar cf - "${proto_files[@]}" \
 | docker run -i pachyderm_proto \
 | tar xf -

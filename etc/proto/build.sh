#!/bin/bash
# This script build protos using the pachyderm_proto image.
# TODO - fix paths with PWD
read -ra proto_files < <(find src -name "*.proto" -print0 | xargs -0)
tar cf - "${proto_files[@]}" \
 | docker run -i --mount type=bind,source="/home/docksonwedge/api-test/",target="/tmp/api-test/" --rm pachyderm_proto \
 | tar xf -

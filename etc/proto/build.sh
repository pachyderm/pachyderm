#!/bin/bash
# This script build protos using the pachyderm_proto image.

find src -regex ".*\.proto" -print0 \
| xargs -0 tar cf - -b20 \
| docker run -i pachyderm_proto \
| tar xf - -b20

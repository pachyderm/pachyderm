#!/bin/bash
# This script build protos using the pachyderm_proto image.

tar cf - $(find src -name "*.proto") $(find etc/proto/pachgen) \
 | docker run -i pachyderm_proto
#tar cf - $(find src -name "*.proto") \
#| docker run -i pachyderm_proto \
#| tar xf -

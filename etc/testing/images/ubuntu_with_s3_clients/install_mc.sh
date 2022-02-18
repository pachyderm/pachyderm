#!/bin/bash
#

OS_ARCH="$( echo "$TARGETPLATFORM" | cut -d/ -f2)"

wget -nv -O /usr/local/bin/mc https://dl.min.io/client/mc/release/linux-"${OS_ARCH}"/mc
chmod ugo+x /usr/local/bin/mc

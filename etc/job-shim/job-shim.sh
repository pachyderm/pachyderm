#!/bin/sh
if [ ! -d "/pfs/.bin" ]; then
    mkdir -p /pfs/.bin
    cp /pach/* /pfs/.bin/
fi

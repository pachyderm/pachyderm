#!/bin/sh

env

find /pfs/books/ -type f -exec /scripts/jsonify.sh {} \;

bleve create /pfs/out/index -m /scripts/mapping.json
bleve index /pfs/out/index /pfs/books/* -d

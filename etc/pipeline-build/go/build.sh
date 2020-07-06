#!/bin/sh
set -e
go version
cd /pfs/source
go build -o /pfs/out/main
cp /app/run /pfs/out/run

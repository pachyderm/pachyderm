#!/bin/sh
set -e
go version
cd /pfs/__source__
go build -o out

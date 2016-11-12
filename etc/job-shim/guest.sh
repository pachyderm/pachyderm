#!/bin/sh
cp /pfs/.bin/mount.fuse /sbin/mount.fuse
cp /pfs/.bin/fusermount /bin/fusermount
/pfs/.bin/job-shim $1

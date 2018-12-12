#!/bin/bash

FILES=*

REPO="trips"

pachctl create-repo $REPO

for f in $FILES
do
	if [ "$f" != "load.sh" ]
	then
		echo $f
		pachctl put-file $REPO master $f -c -f $f
	fi
	sleep 1
done

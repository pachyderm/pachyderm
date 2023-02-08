#!/bin/sh

echo "{}" | jq -c --rawfile content "$1" --arg path "$1" '.path |= $path | .content |= $content' > $1.json
mv $1.json $1

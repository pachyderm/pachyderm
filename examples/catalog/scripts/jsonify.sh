#!/bin/sh

echo "{}" | jq -c --rawfile content "$1" \
				  --arg path $(echo $1 | sed "s/\/pfs\///g") \
				  --arg commit \
				  "$PACH_OUTPUT_COMMIT_ID" '.path |= $path | .commit |= $commit | .content |= $content' > $1.json
mv $1.json $1

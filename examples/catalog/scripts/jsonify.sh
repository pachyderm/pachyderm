#!/bin/sh

echo "{}" | jq -c --rawfile content "$1" \
				  --arg repo $(echo $1 | cut -d'/' -f3) \
				  --arg commit "$PACH_OUTPUT_COMMIT_ID" \
				  --arg path $(echo $1 | cut -d'/' -f4) \
				  '.repo |= $repo | .commit |= $commit | .path |= $path | .content |= $content' > $1.json
mv $1.json $1

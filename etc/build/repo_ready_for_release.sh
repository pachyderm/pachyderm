#!/bin/bash

if git diff-index --quiet HEAD --; then
	# No changes
	VERSION=$("$GOPATH/bin/pachctl" version --client-only)
	if [ -z "$VERSION" ]; then
		echo "Missing version information. Exiting."
		exit 1
	fi
	if git tag | grep -q "$VERSION\$"; then
		echo "Tag $VERSION already exists. Exiting"
		exit 1
	fi
else
	# Changes
	echo "Local changes not committed. Aborting."
	exit 1
fi


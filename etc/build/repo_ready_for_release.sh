#!/bin/bash

if git diff-index --quiet HEAD --; then
	# No changes
	VERSION=`$GOPATH/bin/pachctl version --client-only`
	if [ -z $VERSION ]; then
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

system_version="$(pachctl version --client-only)"
built_version="$(${GOPATH}/bin/pachctl version --client-only)"
if [[ "${system_version}" != "${built_version}" ]]; then
  echo "'pachctl version' disagrees with '\${GOPATH}/bin/pachctl version'"
  echo "pachctl version:               \"${system_version}\""
  echo "\${GOPATH}/bin/pachctl version: \"${built_version}\""
  echo "'pachctl version' disagrees with '${GOPATH}/bin/pachctl version'"
  exit 1
fi

#!/bin/bash
#
# This builds a version of the pachd and worker images based on Red Hat's UBI,
# rather than on 'scratch'.

set -ex


# Validate env vars
if [[ -z "${CIRCLE_TAG}" ]]; then
  echo "Must set CIRCLE_TAG to release tag (e.g. '2.1.20')" >/dev/stderr
  exit 1
elif [[ -z "${REDHAT_MARKETPLACE_CONSOLE_OSPID}" ]]; then
  echo "Must set REDHAT_MARKETPLACE_CONSOLE_OSPID" >/dev/stderr
  exit 1
elif [[ "${REDHAT_MARKETPLACE_CONSOLE_OSPID}" != ospid-* ]]; then
  echo "Must set REDHAT_MARKETPLACE_CONSOLE_OSPID to ospid-* (with prefix)" >/dev/stderr
  exit 1
elif [[ -z "${REDHAT_MARKETPLACE_CONSOLE_PASSWORD}" ]]; then
  echo "Must set REDHAT_MARKETPLACE_CONSOLE_PASSWORD" >/dev/stderr
  exit 1
fi

# abbreviate name, for convenience
ospid="${REDHAT_MARKETPLACE_CONSOLE_OSPID}"
version="${CIRCLE_TAG}"  # designed to run in CircleCI on tags (releases)

# Build console image based on RedHat's UBI instead of bog standard node image
cd "$(git rev-parse --show-toplevel)"
docker build . -t "scan.connect.redhat.com/${ospid}/console:${version}" -f etc/redhat/Dockerfile.redhat

# Push the image to our Red Hat Technology Portal project
docker login -u unused scan.connect.redhat.com --password-stdin <<<"${REDHAT_MARKETPLACE_CONSOLE_PASSWORD}"
docker push "scan.connect.redhat.com/${ospid}/console:${version}"

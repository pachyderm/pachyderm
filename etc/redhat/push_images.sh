#!/bin/bash
#
# This builds a version of the pachd and worker images based on Red Hat's UBI,
# rather than on 'scratch'.

set -ex

# Validate env vars
if [[ -z "${CIRCLE_TAG}" ]]; then
  echo "Must set CIRCLE_TAG to release tag (e.g. '2.1.20')" >/dev/stderr
  exit 1
fi
for img in pachd worker; do
  echo "checking REDHAT_MARKETPLACE_${img^^}_OSPID REDHAT_MARKETPLACE_${img^^}_PASSWORD"
  ospid=$(eval "echo \${REDHAT_MARKETPLACE_${img^^}_OSPID}")
  password=$(eval "echo \${REDHAT_MARKETPLACE_${img^^}_PASSWORD}")
  if [[ -z "${ospid}" ]]; then
    echo "Must set REDHAT_MARKETPLACE_${img^^}_OSPID" >/dev/stderr
    exit 1
  elif [[ "${ospid}" != ospid-* ]]; then
    echo "Must set REDHAT_MARKETPLACE_${img^^}_OSPID to ospid-* (with prefix)" >/dev/stderr
    exit 1
  elif [[ -z "${password}" ]]; then
    echo "Must set REDHAT_MARKETPLACE_${img^^}_PASSWORD" >/dev/stderr
    exit 1
  fi
done

# Generate Dockerfiles based on RedHat's UBI instead of scratch.
#
# This is effectively required for our RedHat Marketplace offering (as
# implemented by our OpenShift operator in
# github.com/pachyderm/openshift-operator. Users are given an interface that
# allows them to run a shell in their pachd/worker containers, and RedHat's
# Universal Base Image provides a shell they can run while conforming to the
# the RedHat Marketplace image approval rules)
for img in pachd worker; do
  sed \
    's#FROM scratch#FROM registry.access.redhat.com/ubi8/ubi-minimal#g' \
    "Dockerfile.${img}" \
    >"Dockerfile.redhat.${img}"
done

# Build images from the modified Dockerfiles
make docker-build

# Push the image to our Red Hat Technology Portal project
for img in pachd worker; do
  ospid=$(eval "echo \${REDHAT_MARKETPLACE_${img^^}_OSPID}")
  password=$(eval "echo \${REDHAT_MARKETPLACE_${img^^}_PASSWORD}")
  docker login -u unused scan.connect.redhat.com --password-stdin <<<"${password}"
  docker tag "pachyderm/${img}:local" "scan.connect.redhat.com/${ospid}/${img}:${CIRCLE_TAG}"
  docker push "scan.connect.redhat.com/${ospid}/${img}:${CIRCLE_TAG}"
done

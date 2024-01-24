#!/bin/bash

set -euo pipefail

commit_sha=$(git rev-parse HEAD)
echo "COMMIT_SHA $commit_sha"

git_branch=$(git rev-parse --abbrev-ref HEAD)
echo "GIT_BRANCH $git_branch"

git_tree_status=$(git diff-index --quiet HEAD -- && echo 'Clean' || echo 'Modified')
echo "GIT_TREE_STATUS $git_tree_status"

most_recent_tag=$(git tag -l 'v*' --sort=creatordate | tail -n1)
echo "STABLE_VERSION $most_recent_tag"

# TODO(jrockway): Read these from the git tags when the branch is named "2.x".
echo "STABLE_APP_VERSION 2.9.0"

additional_version=$(git diff-index --quiet HEAD -- && echo "-${commit_sha}" || echo "-${commit_sha}+dirty")
echo "STABLE_ADDITIONAL_VERSION $additional_version"

ci_runner_image_version="$(date +%Y%m%d)-${commit_sha}"
echo "STABLE_CI_RUNNER_IMAGE_VERSION ${ci_runner_image_version}"

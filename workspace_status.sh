#!/bin/bash

set -euo pipefail

commit_sha=$(git rev-parse HEAD)
echo "COMMIT_SHA $commit_sha"

git_branch=$(git rev-parse --abbrev-ref HEAD)
echo "GIT_BRANCH $git_branch"

git_tree_status=$(git diff-index --quiet HEAD -- && echo 'Clean' || echo 'Modified')
echo "GIT_TREE_STATUS $git_tree_status"

exact=$(git describe --exact-match 2>/dev/null | sed -e s/^v// | cut -d - -f 1 || echo 2.10.0)
echo "STABLE_APP_VERSION $exact"

additional_version=$(git describe --exact-match 2>/dev/null | cut -d - -f 2- || echo "pre.$(git describe --long --dirty=x | rev | cut -d - -f 1 | rev)")
echo "STABLE_ADDITIONAL_VERSION -$additional_version"

ci_runner_image_version="$(date +%Y%m%d)-${commit_sha}"
echo "STABLE_CI_RUNNER_IMAGE_VERSION ${ci_runner_image_version}"

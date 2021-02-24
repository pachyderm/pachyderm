#!/bin/sh

set -ve

find ./src -name '*.go' -type f -exec sed -i '' 's#github.com/pachyderm/pachyderm/#github.com/pachyderm/pachyderm/v2/#g' {} \;
find ./src -name '*.proto' -type f -exec sed -i '' 's#github.com/pachyderm/pachyderm/#github.com/pachyderm/pachyderm/v2/#g' {} \;
go mod edit -module "github.com/pachyderm/pachyderm/v2"

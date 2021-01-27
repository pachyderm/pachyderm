#!/bin/sh

set -ve

find ./src -name '*.go' -type f -exec sed -i '' 's#github.com/pachyderm/pachyderm/src/server/pkg/#github.com/pachyderm/pachyderm/src/internal/#g' {} \;
find ./src -name '*.go' -type f -exec sed -i '' 's#github.com/pachyderm/pachyderm/src/client/pkg/#github.com/pachyderm/pachyderm/src/internal/#g' {} \;

find ./src -name '*.go' -type f -exec sed -i '' 's#github.com/pachyderm/pachyderm/src/client/pfs#github.com/pachyderm/pachyderm/src/pfs#g' {} \;
find ./src -name '*.go' -type f -exec sed -i '' 's#github.com/pachyderm/pachyderm/src/client/pps#github.com/pachyderm/pachyderm/src/pps#g' {} \;
find ./src -name '*.go' -type f -exec sed -i '' 's#github.com/pachyderm/pachyderm/src/client/auth#github.com/pachyderm/pachyderm/src/auth#g' {} \;
find ./src -name '*.go' -type f -exec sed -i '' 's#github.com/pachyderm/pachyderm/src/client/admin#github.com/pachyderm/pachyderm/src/admin#g' {} \;
find ./src -name '*.go' -type f -exec sed -i '' 's#github.com/pachyderm/pachyderm/src/client/debug#github.com/pachyderm/pachyderm/src/debug#g' {} \;
find ./src -name '*.go' -type f -exec sed -i '' 's#github.com/pachyderm/pachyderm/src/client/enterprise#github.com/pachyderm/pachyderm/src/enterprise#g' {} \;
find ./src -name '*.go' -type f -exec sed -i '' 's#github.com/pachyderm/pachyderm/src/client/health#github.com/pachyderm/pachyderm/src/health#g' {} \;
find ./src -name '*.go' -type f -exec sed -i '' 's#github.com/pachyderm/pachyderm/src/client/transaction#github.com/pachyderm/pachyderm/src/transaction#g' {} \;

find ./src -name '*.proto' -type f -exec sed -i '' 's#github.com/pachyderm/pachyderm/src/client/pfs#github.com/pachyderm/pachyderm/src/pfs#g' {} \;
find ./src -name '*.proto' -type f -exec sed -i '' 's#github.com/pachyderm/pachyderm/src/client/pps#github.com/pachyderm/pachyderm/src/pps#g' {} \;
find ./src -name '*.proto' -type f -exec sed -i '' 's#github.com/pachyderm/pachyderm/src/client/auth#github.com/pachyderm/pachyderm/src/auth#g' {} \;
find ./src -name '*.proto' -type f -exec sed -i '' 's#github.com/pachyderm/pachyderm/src/client/admin#github.com/pachyderm/pachyderm/src/admin#g' {} \;
find ./src -name '*.proto' -type f -exec sed -i '' 's#github.com/pachyderm/pachyderm/src/client/debug#github.com/pachyderm/pachyderm/src/debug#g' {} \;
find ./src -name '*.proto' -type f -exec sed -i '' 's#github.com/pachyderm/pachyderm/src/client/enterprise#github.com/pachyderm/pachyderm/src/enterprise#g' {} \;
find ./src -name '*.proto' -type f -exec sed -i '' 's#github.com/pachyderm/pachyderm/src/client/health#github.com/pachyderm/pachyderm/src/health#g' {} \;
find ./src -name '*.proto' -type f -exec sed -i '' 's#github.com/pachyderm/pachyderm/src/client/transaction#github.com/pachyderm/pachyderm/src/transaction#g' {} \;

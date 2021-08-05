#!/bin/bash

set -euxo pipefail

wget https://github.com/pachyderm/hubcli/releases/download/v0.0.1-beta.1/hubcli
chmod a+x hubcli
./hubcli --endpoint https://hub.pachyderm.com/api/graphql --apikey $HUB_API_KEY --op create-workspace-and-wait --orgid 2193 --loglevel trace --infofile workspace.json
./hubcli --endpoint https://hub.pachyderm.com/api/graphql --apikey $HUB_API_KEY --op delete-workspace --loglevel trace --infofile workspace.json

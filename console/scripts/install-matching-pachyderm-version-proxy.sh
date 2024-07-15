#!/bin/bash

helm repo add pachyderm https://pachyderm.github.io/helmchart 
helm repo update

helm uninstall pachyderm

PACHYDERM_VERSION=$(jq -r .pachyderm ./version.json)
RELEASE_COMMIT=$(jq -r .pachReleaseCommit ./version.json)
echo Installing version $PACHYDERM_VERSION with pachd.image.tag $RELEASE_COMMIT

helm install \
	--wait --timeout 10m pachyderm pachyderm/pachyderm \
	--version=$PACHYDERM_VERSION \
	--set deployTarget=LOCAL \
	--set console.enabled=false \
	--set console.config.disableTelemetry=true \
    --set proxy.enabled=true \
    --set proxy.host=localhost \
    --set pachd.metrics.enabled=false \
    --set proxy.service.type=LoadBalancer

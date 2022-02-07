#!/bin/bash

export PACHYDERM_VERSION="$(jq -r .pachyderm version.json)"

pachctl create repo images
pachctl put file images@master:liberty.png -f http://imgur.com/46Q8nDz.png
pachctl create pipeline -f https://raw.githubusercontent.com/pachyderm/pachyderm/v${PACHYDERM_VERSION}/examples/opencv/edges.json
pachctl auth set repo images repoReader user:john-doe@pachyderm.io

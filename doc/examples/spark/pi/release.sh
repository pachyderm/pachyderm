#! /bin/bash

USAGE="Usage: release.sh tag-name"

if [ "$#" -ne 1 ]; then
  echo "Exaclty one parameter is needed"
  echo $USAGE
  exit -1
fi

TAGNAME="$1"

docker build -t pachyderm/estimate-pi-spark:${TAGNAME} -t pachyderm/estimate-pi-spark:latest .
docker push pachyderm/estimate-pi-spark:${TAGNAME}
docker push pachyderm/estimate-pi-spark:latest

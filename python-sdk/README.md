# Python-SDK

This directory contains the auto-generated python SDK for interacting with 
  a Pachyderm cluster. This currently uses a fork of betterproto that can
  be found here: https://github.com/bbonenfant/python-betterproto/tree/grpcio-support


## Notes
To rebuild docker image (cd pachyderm/python-sdk)
```bash
docker build -t pachyderm_python_proto:python-sdk .
```

To regenerate proto files (cd pachyderm)
```bash
rm -rf python-sdk/pachyderm_sdk/api
find src -regex ".*\.proto" \
  | grep -v 'internal' \
  | grep -v 'server' \
  | xargs tar cf - \
  | docker run -i pachyderm_python_proto:python-sdk \
  | tar -C python-sdk/pachyderm_sdk -xf -
```
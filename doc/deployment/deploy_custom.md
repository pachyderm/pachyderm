# Custom Deployments

Coming soon.
This document, when complete, will detail the various options of the `pachctl deploy custom ...` command for an on-premises deployment.

## Creating a Pachyderm manifest


```sh

pachctl deploy custom --persistent-disk google --object-store s3 foo 100 pach-minio-bucket OBSIJRBE0PP2NO4QOA27  tfteSlswRu7BJ86wekitnifILbZam1KYY3TG  localhost:9000 --static-etcd-volume=foo --local-roles --dry-run > local_roles.json


pachctl deploy custom --persistent-disk google --object-store s3 <persistent disk name> <persistent disk size> <object store bucket> <object store id> <object store secret> <object store endpoint> --static-etcd-volume=${STORAGE_NAME} --dry-run > deployment.json
```

Then you can modify `deployment.json` to fit your environment and kubernetes deployment.  Once, you have your manifest ready, deploying Pachyderm is as simple as:

```sh
kubectl create -f deployment.json
```

## Editing a Pachyderm manifest


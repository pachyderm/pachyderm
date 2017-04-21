# On Premises

Pachyderm is built on [Kubernetes](https://kubernetes.io/) and can be backed by an object store of your choice. As such, Pachyderm can run on any on premise platforms/frameworks that support Kubernetes, a persistent disk/volume, and an object store.

## Prerequisites

1. [kubectl](https://kubernetes.io/docs/user-guide/prereqs/)
2. [pachctl](http://docs.pachyderm.io/en/latest/pachctl/pachctl.html)

## Kubernetes

The Kubernetes docs have instructions for [deploying Kubernetes in a variety of on-premise scenarios](https://kubernetes.io/docs/getting-started-guides/#on-premises-vms).  We recommend following one of these guides to get Kubernetes running on premise.

## Object Store

Once you have Kubernetes up and running, deploying Pachyderm is a matter of supplying Kubernetes with a JSON/yaml manifest to create the Pachyderm resources.  This includes providing information that Pachyderm will use to connect to a backing object store.

For on premise deployments, we recommend using [Minio](https://minio.io/) as a backing object store.  However, at this point, you could utilize any backing object store that has an S3 compatible API.  To create a manifest template for your on premise deployment, run:

```sh
pachctl deploy custom --persistent-disk google --object-store s3 <persistent disk name> <persistent disk size> <object store bucket> <object store id> <object store secret> <object store endpoint> --static-etcd-volume=${STORAGE_NAME} --dry-run > deployment.json
```

Then you can modify `deployment.json` to fit your environment and kubernetes deployment.  Once, you have your manifest ready, deploying Pachyderm is as simple as:

```sh
kubectl create -f deployment.json
```

## Need Help?

If you need help with your on premises deploy, please reach out to us on Pachyderm's [slack channel](https://pachyderm-users.slack.com/messages) or via email at support@pachyderm.io. We are happy to help!


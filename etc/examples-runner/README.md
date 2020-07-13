# Examples runner

This is a docker image for running the examples end-to-end test script.

The image is available on docker hub `pachyderm/examples-runner`. For now, no
versioning is applied - it'll just use the latest version of th examples
script on master, as well as the latest pachctl available in
`pachyderm/pachctl:latest`.

The runner script takes an optional parameter specifying the cluster pachd
address to run on. To, e.g., run against minikube, this should work:

```bash
docker run -it pachyderm/examples-runner $(minikube ip):30650
```

If no parameter is specified, the examples runner will default to using
`$PACHD_PEER_SERVICE_HOST:$PACHD_PEER_SERVICE_PORT` if those env vars or
available, or otherwise `localhost:30650`.

You can also run this directly in a kubernetes cluster if it has pachyderm
deployed, like so:

```bash
kubectl run -it examples-runner --image=pachyderm/examples-runner
```

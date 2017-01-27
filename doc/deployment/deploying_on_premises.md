# Deploying on premises

.. note::

	The following on premises deployment options are experimental and the related deploy commands will likely change in the next release. We'd love your feedback an how you're using it and what we can improve. info@pachyderm.io.

This guide will walk you through the recommended current recommended path to test an "on premises" Pachyderm deployment. This will illustrate the use of a privately managed and hosted object store (minio) with a pachyderm cluster.  However, for the time being, the following deployment is not scalable.  It is hear for illustrative purposes.

## Prerequisites
- [Minikube](#minikube) (and VirtualBox)
- [Pachyderm Command Line Interface](#pachctl)
- A locally running instance of [Minio](https://www.minio.io/) (e.g., via Docker)

## Deploy Pachyderm with a Minio Backend
Assuming that you have Minikube running, it's incredibly easy to deploy Pachyderm backed by a locally running Minio object store at `127.0.0.1:9000`.  

```sh
pachctl deploy minio <id> <secret> 10.0.2.2:9000
```

This generates a Pachyderm manifest and deploys Pachyderm on Kubernetes. It also instructs Pachyderm to use the Minio instance for the storage backend (note `10.0.2.2` is used here instead of `127.0.0.1`, because minikube is running inside of virtual box).  It may take a few minutes for the pachd nodes to be running because it's pulling containers from DockerHub. You can see the cluster status by using `kubectl get all`.

## Port Forwarding

The last step is to set up port forwarding so commands you send can reach Pachyderm within the VM. We background this process since port forwarding blocks.

```shell
$ pachctl port-forward &
```

Once port forwarding is complete, pachctl should automatically be connected. Try `pachctl version` to make sure everything is working.

```shell
$ pachctl version
COMPONENT           VERSION
pachctl             1.3.4
pachd               1.3.4
```

We're good to go!

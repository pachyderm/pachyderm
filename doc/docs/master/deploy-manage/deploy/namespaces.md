# Non-Default Namespaces

Often, production deploys of Pachyderm involve deploying Pachyderm to a non-default namespace. This helps administrators of the cluster more easily manage Pachyderm components alongside other things that might be running inside of Kubernetes (DataDog, TensorFlow Serving, etc.).

To deploy Pachyderm to a non-default namespace, you just need to add the `-n` or `--namespace` flag when deploying. 
If the namespace doesn't already exist, you can have helm create it with `--create-namespace`:

```shell
helm install <args> --namespace pachyderm --create-namespace
```

To talk to your Pachyderm cluster, you can either modify an existing pachctl context
```shell
pachctl config update context --namespace pachyderm
```

or import one from Kubernetes:
```shell
pachctl config import-kube new-context-name --namespace pachyderm
```


After the Pachyderm pods are up and running, you should see a pod for `pachd` running
(alongside etcd, pg-bouncer or postgres, console, depending on your installation):

```shell
kubectl config set-context $(kubectl config current-context) --namespace=pachyderm
kubectl get pods
```

**System Response:**

```
NAME                     READY     STATUS    RESTARTS   AGE
pachd-784bdf7cd7-7dzxr   1/1       Running   0          3m
...
```

!!! Note
    If you want to use RBAC for multiple Pachyderm clusters in different namespaces within the same cluster,
    you should disable cluster-level roles by setting the value of `pachd.rbac.clusterRBAC` to `false` in your helm chart
    or passing `--set pachd.rbac.clusterRBAC=false` when deploying.

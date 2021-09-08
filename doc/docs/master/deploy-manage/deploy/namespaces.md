# Non-Default Namespaces

Often, production deploys of Pachyderm involve deploying Pachyderm to a non-default namespace. This helps administrators of the cluster more easily manage Pachyderm components alongside other things that might be running inside of Kubernetes (DataDog, TensorFlow Serving, etc.).

To deploy Pachyderm to a non-default namespace, you just need to create that namespace with `kubectl` and then add the `--namespace` flag to your deploy command:

```
kubectl create namespace pachyderm
kubectl config set-context $(kubectl config current-context) --namespace=pachyderm
pachctl deploy <args> --namespace pachyderm
```

After the Pachyderm pods are up and running, you should see something similar to:

```shell
kubectl get pods
```

Once the pods are up, you should see a pod for `pachd` running 
(alongside etcd, pg-bouncer or postgres, console, depending on your installation). 

**System Response:**

```
NAME                     READY     STATUS    RESTARTS   AGE
pachd-784bdf7cd7-7dzxr   1/1       Running   0          3m
...
```

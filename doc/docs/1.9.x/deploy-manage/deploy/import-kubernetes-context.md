# Import a Kubernetes Context

**Note:** The steps in this section apply to your configuration only
if you deployed Pachyderm from a manifest created by the `pachctl deploy`
command with the `--dry-run` flag. If you did not use the `--dry-run` flag,
skip this section.

When you run the `pachctl deploy` command with `--dry-run` flag, instead of
immediately deploying a cluster, the command creates a Kubernetes
deployment manifest that you can further edit and later use to deploy a
Pachyderm cluster.

You can use that manifest with a standard `kubectl apply` command to deploy
Pachyderm. For example, if you have created a manifest called
`test-manifest.yaml`, you can deploy a Pachyderm cluster by running the
following command:

```shell
kubectl apply -f test-manifest.yaml
```

Typically, when you run `pachctl deploy`,
Pachyderm creates a new Pachyderm context with the information from
the current
[Kubernetes context](https://kubernetes.io/docs/concepts/configuration/organize-cluster-access-kubeconfig/#context)
embedded into it.

When you use the `--dry-run` flag, the Pachyderm context is not created.
Therefore, if you deploy a Pachyderm cluster from a manifest that you have
created earlier, you need to manually create a new Pachyderm context with
the embedded current Kubernetes context and activate that context.

To import a Kubernetes context, complete the following steps:

1. Deploy a Pachyderm cluster from the Kubernetes manifest that you have
   created when you ran the `pachctl deploy` command with the `--dry-run`
   flag:

   ```shell
   $ kubectl apply -f <manifest.yaml>
   clusterrole.rbac.authorization.k8s.io/pachyderm configured
   clusterrolebinding.rbac.authorization.k8s.io/pachyderm configured
   deployment.apps/etcd configured
   service/etcd configured
   service/pachd configured
   deployment.apps/pachd configured
   service/dash configured
   deployment.apps/dash configured
   secret/pachyderm-storage-secret configured
   ```

1. Verify that the cluster was successfully deployed:

   ```shell
   $ kubectl get pods
   NAME                     READY   STATUS    RESTARTS   AGE
   dash-64c868cc8b-j79d6    2/2     Running   0          20h
   etcd-6865455568-tm5tf    1/1     Running   0          20h
   pachd-6464d985c7-dqgzg   1/1     Running   0          70s
   ```

   You must see all the `dash`, `etcd`, and `pachd` pods running.

1. Create a new Pachyderm context with the embedded Kubernetes context:

   ```shell
   $ pachctl config set context <new-pachyderm-context> -k `kubectl config current-context`
   ```

1. Verify that the context was successfully created and view the context parameters:

   **Example:**

   ```shell
   $ pachctl config get context test-context
   {
     "source": "IMPORTED",
     "cluster_name": "minikube",
     "auth_info": "minikube",
     "namespace": "default"
   }
   ```

1. Activate the new Pachyderm context:

   ```shell
   $ pachctl config set active-context <new-pachyderm-context>
   ```

1. Verify that the new context has been activated:

   ```shell
   $ pachctl config get active-context
   ```

# Import a Kubernetes Context

After you've deployed Pachyderm with helm, the Pachyderm context is not created.
Therefore, you need to manually create a new Pachyderm context with
the embedded current Kubernetes context and activate that context.

To import a Kubernetes context, complete the following steps:

1. Deploy a Pachyderm cluster using the helm installation commands.

1. Verify that the cluster was successfully deployed:

   ```shell
   kubectl get pods
   ```

   **System Response:**

   ```shell
   NAME                     READY   STATUS    RESTARTS   AGE
   dash-64c868cc8b-j79d6    2/2     Running   0          20h
   etcd-6865455568-tm5tf    1/1     Running   0          20h
   pachd-6464d985c7-dqgzg   1/1     Running   0          70s
   ```

   You must see all the `dash`, `etcd`, and `pachd` pods running.

1. Create a new Pachyderm context with the embedded Kubernetes context:

   ```shell
   pachctl config import-kube <new-pachyderm-context> -k `kubectl config current-context`
   ```

1. Verify that the context was successfully created and view the context parameters:

   **Example:**

   ```shell
   pachctl config get context test-context
   ```

   **System Response:**

   ```shell
   {
     "source": "IMPORTED",
     "cluster_name": "minikube",
     "auth_info": "minikube",
     "namespace": "default"
   }
   ```

1. Activate the new Pachyderm context:

   ```shell
   pachctl config set active-context <new-pachyderm-context>
   ```

1. Verify that the new context has been activated:

   ```shell
   pachctl config get active-context
   ```

# Non-Default Namespaces

Often, production deploys of Pachyderm involve deploying Pachyderm to a non-default namespace. This helps administrators of the cluster more easily manage Pachyderm components alongside other things that might be running inside of Kubernetes (DataDog, TensorFlow Serving, etc.).

* To deploy Pachyderm to a non-default namespace, 
you need to add the `-n` or `--namespace` flag when deploying. 
    If the namespace does not already exist, 
    you can have [Helm](../helm-install/) create it with `--create-namespace`.


    ```shell
    helm install <args> --namespace pachyderm --create-namespace
    ```

* To talk to your Pachyderm cluster:

    - You can either modify an existing pachctl context
        ```shell
        pachctl config update context --namespace pachyderm
        ```

    - or [import one from Kubernetes](../import-kubernetes-context/):

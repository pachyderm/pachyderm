# Using Examples to enable RBAC

This guide helps in setting up RBAC for Kubeflow.

The RBAC rules here assume 3 groups: admin, datascience and validator as sample groups for operating on Kubeflow.

## Setup

```
./apply_example.sh --issuer https://dex.example.com:32000 --jwks-uri https://dex.example.com:32000/keys --client-id ldapdexapp
```

### Note Regarding Istio RBAC

Currently, the only service authenticated and authorized supported in this example is ml-pipeline service.
Support for authorization in Pipelines is being discussed in this [issue](https://github.com/kubeflow/pipelines/issues/1223).
This example allows for authentication and authorization only for requests within the Kubernetes cluster.

# Configure `pachd` Environment Variables

The environment (env) variables in this article define the parameters for your Pachyderm daemon container. 

You can reference env variables in your user code. For example, if your code writes data to an external system and you want to know the current job ID, you can use the `PACH_JOB_ID` environment variable to refer to the current job ID.

## View Current Env Variables

To list your current `pachd` env variables, use the following command: 

```shell
kubectl get deploy pachd -o yaml
```

---

## Variables

| Variable | Required | Type | Default | Description |
|---|---|---|---|---|
| STORAGE_BACKEND | Yes | string |  | The backend storage solution; This is set automatically if deployTarget is `GOOGLE`, `AMAZON`, `MICROSOFT`, or `LOCAL`. |
| STORAGE_HOST_PATH | No | string |  | The storage host's path.  |
| PFS_ETCD_PREFIX | No | string | pachyderm_pfs |  |
| KUBERNETES_PORT_443_TCP_ADDR | Yes | string |  |  |
| INIT | No | boolean | FALSE |  |
| WORKER_IMAGE | No | string |  | The worker's image. |
| WORKER_SIDECAR_IMAGE | No | string |  | The worker sidecar's image. |
| WORKER_IMAGE_PULL_POLICY | No | string |  | The pull policy for the worker's image. |
| IMAGE_PULL_SECRETS | No | string |  | The pull secrets for the image. |
| PACHD_MEMORY_REQUEST | No | string | 1T | The amount of memory requested for `pachd`. |
| WORKER_USES_ROOT | No | boolean | FALSE | The option to allow the worker to use root privileges. |
| REQUIRE_CRITICAL_SERVERS_ONLY | No | boolean | FALSE | The option to only require critical servers. |
| PACHD_POD_NAME | Yes | string |  | The name of the `pachd` pod.  |
| ENABLE_WORKER_SECURITY_CONTEXTS | No | boolean | TRUE | The option to enable security contexts for workers. |
| TLS_CERT_SECRET_NAME | No | string |  | The secret name for a TLS certification. |
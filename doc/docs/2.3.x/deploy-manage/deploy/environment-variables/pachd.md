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

### Global 

| Environment Variable   | Default Value     | Description |
| ---------------------- | ----------------- | ----------- |
| `ETCD_SERVICE_HOST`    | N/A               | The host on which the etcd service runs. |
| `ETCD_SERVICE_PORT`    | N/A               | The etcd port number.                    |
| `PPS_WORKER_GRPC_PORT` | `80`              | The GRPs port number.                    |
| `PORT`                 | `650`             | The `pachd` port number. |
| `HTTP_PORT`             | `652`             | The HTTP port number.   |
| `PEER_PORT`             | `653`             | The port for pachd-to-pachd communication. |
| `NAMESPACE`            | `deafult`         | The namespace in which Pachyderm is deployed. |

### General 

| Environment Variable       | Default Value | Description |
| -------------------------- | ------------- | ----------- |
| `NUM_SHARDS`               | `32`      | The max number of `pachd` pods that can run in a <br> single cluster. |
| `STORAGE_BACKEND`          | `""`      | The storage backend defined for the Pachyderm cluster.|
| `STORAGE_HOST_PATH`        | `""`      | The host path to storage. |
| `KUBERNETES_PORT_443_TCP_ADDR` |`none` | An IP address that Kubernetes exports <br> automatically for your code to communicate with <br> the Kubernetes API. Read access only. Most variables <br> that have use the `PORT_ADDRESS_TCP_ADDR` pattern <br> are Kubernetes environment variables. For more information,<br> see [Kubernetes environment variables](https://kubernetes.io/docs/concepts/services-networking/service/#environment-variables){target=_blank}. |
| `METRICS`                  | `true`   | Defines whether anonymous Pachyderm metrics are being <br>collected or not. |
| `BLOCK_CACHE_BYTES`        | `1G`     | The size of the block cache in `pachd`. |
| `WORKER_IMAGE`             | `""`     | The base Docker image that is used to run your pipeline.|
| `WORKER_SIDECAR_IMAGE`     | `""`     | The `pachd` image that is used as a worker sidecar. |
| `WORKER_IMAGE_PULL_POLICY` | `IfNotPresent`| The pull policy that defines how Docker images are <br>pulled. You can set <br> a Kubernetes image pull policy as needed. |
| `LOG_LEVEL`                | `info`   | Verbosity of the log output. If you want to disable <br> logging, set this variable to `0`. Viable Options <br>`debug` <br>`info` <br> `error`<br>For more information, see [Go logrus log levels](https://pkg.go.dev/github.com/sirupsen/logrus#Level){target=_blank}. ||
| `IAM_ROLE`                 |  `""`    | The role that defines permissions for Pachyderm in AWS.|
| `IMAGE_PULL_SECRET`        |  `""`    | The Kubernetes secret for image pull credentials.|
| `EXPOSE_OBJECT_API`        |  `false` | Controls access to internal Pachyderm API.|
| `WORKER_USES_ROOT`         |  `true`  | Controls root access in the worker container.|
| `S3GATEWAY_PORT`           |  `600`   | The S3 gateway port number|
| `DISABLE_COMMIT_PROGRESS_COUNTER` |`false`| A feature flag that disables commit propagation <br> progress counter. If you have a large DAG, <br> setting this parameter to `true` might help <br> improve etcd performance. You only need to set <br>this parameter on the `pachd` pod. Pachyderm passes <br> this parameter to worker containers automatically. |

### Storage 

| Environment Variable       | Default Value     | Description |
| -------------------------- | ----------------- | ----------- |
| `STORAGE_MEMORY_THRESHOLD` | N/A               | Defines the storage memory threshold. |
| `STORAGE_SHARD_THRESHOLD`  | N/A               | Defines the storage shard threshold.  |
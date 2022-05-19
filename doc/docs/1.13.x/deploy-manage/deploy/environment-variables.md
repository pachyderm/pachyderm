# Configure Environment Variables

When you use Pachyderm, you can define environment variables that
can transmit the required configuration directly to your application.

In Pachyderm, you can define the following types of environment
variables:

* `pachd` environment variables that define parameters for your
Pachyderm daemon container.

* Pachyderm worker environment variables that define parameters
on the Kubernetes pods that run your pipeline code.

You can reference environment variables in your code. For example,
if your code writes data to an external system and you want
to know the current job ID, you can use the `PACH_JOB_ID`
environment variable to refer to the current job ID.

You can access all the variables in the Pachyderm manifest that
is generated when you run `pachctl deploy` with the --dry-run`
flag.

!!! note "See Also:"
    [Deploy Pachyderm](../../../getting-started/local-installation/#deploy-pachyderm)

## `pachd` Environment Variables

You can find the list of `pachd` environment variables in the
`pachd` manifest by running the following command:

```shell
kubectl get deploy pachd -o yaml
```

The following tables list all the `pachd`
environment variables.

**Global Configuration**

| Environment Variable   | Default Value     | Description |
| ---------------------- | ----------------- | ----------- |
| `ETCD_SERVICE_HOST`    | N/A               | The host on which the etcd service runs. |
| `ETCD_SERVICE_PORT`    | N/A               | The etcd port number.                    |
| `PPS_WORKER_GRPC_PORT` | `80`              | The GRPs port number.                    |
| `PORT`                 | `650`             | The `pachd` port number. |
| `HTTP_PORT`             | `652`             | The HTTP port number.   |
| `PEER_PORT`             | `653`             | The port for pachd-to-pachd communication. |
| `NAMESPACE`            | `deafult`         | The namespace in which Pachyderm is deployed. |

**pachd Configuration**

| Environment Variable       | Default Value | Description |
| -------------------------- | ------------- | ----------- |
| `NUM_SHARDS`               | `32`      | The max number of `pachd` pods that can run in a <br> single cluster. |
| `STORAGE_BACKEND`          | `""`      | The storage backend defined for the Pachyderm cluster.|
| `STORAGE_HOST_PATH`        | `""`      | The host path to storage. |
| `KUBERNETES_PORT_443_TCP_ADDR` |`none` | An IP address that Kubernetes exports <br> automatically for your code to communicate with <br> the Kubernetes API. Read access only. Most variables <br> that have use the `PORT_ADDRESS_TCP_ADDR` pattern <br> are Kubernetes environment variables. For more information,<br> see [Kubernetes environment variables](https://kubernetes.io/docs/concepts/services-networking/service/#environment-variables). |
| `METRICS`                  | `true`   | Defines whether anonymous Pachyderm metrics are being <br>collected or not. |
| `BLOCK_CACHE_BYTES`        | `1G`     | The size of the block cache in `pachd`. |
| `WORKER_IMAGE`             | `""`     | The base Docker image that is used to run your pipeline.|
| `WORKER_SIDECAR_IMAGE`     | `""`     | The `pachd` image that is used as a worker sidecar. |
| `WORKER_IMAGE_PULL_POLICY` | `IfNotPresent`| The pull policy that defines how Docker images are <br>pulled. You can set <br> a Kubernetes image pull policy as needed. |
| `LOG_LEVEL`                | `info`   | Verbosity of the log output. If you want to disable <br> logging, set this variable to `0`. Viable Options <br>`debug` <br>`info` <br> `error`<br>For more information, see [Go logrus log levels](https://pkg.go.dev/github.com/sirupsen/logrus#Level). ||
| `IAM_ROLE`                 |  `""`    | The role that defines permissions for Pachyderm in AWS.|
| `IMAGE_PULL_SECRET`        |  `""`    | The Kubernetes secret for image pull credentials.|
| `NO_EXPOSE_DOCKER_SOCKET`  |  `false` | Controls whether you can build images using <br> the `--build` command.|
| `EXPOSE_OBJECT_API`        |  `false` | Controls access to internal Pachyderm API.|
| `WORKER_USES_ROOT`         |  `true`  | Controls root access in the worker container.|
| `S3GATEWAY_PORT`           |  `600`   | The S3 gateway port number|
| `DISABLE_COMMIT_PROGRESS_COUNTER` |`false`| A feature flag that disables commit propagation <br> progress counter. If you have a large DAG, <br> setting this parameter to `true` might help <br> improve etcd performance. You only need to set <br>this parameter on the `pachd` pod. Pachyderm passes <br> this parameter to worker containers automatically. |

**Storage Configuration**

| Environment Variable       | Default Value     | Description |
| -------------------------- | ----------------- | ----------- |
| `STORAGE_MEMORY_THRESHOLD` | N/A               | Defines the storage memory threshold. |
| `STORAGE_SHARD_THRESHOLD`  | N/A               | Defines the storage shard threshold.  |

## Pipeline Worker Environment Variables

Pachyderm defines many environment variables for each Pachyderm
worker that runs your pipeline code. You can print the list
of environment variables into your Pachyderm logs by including
the `env` command into your pipeline specification. For example,
if you have an `images` repository, you can configure your pipeline
specification like this:

```json
{
    "pipeline": {
        "name": "env"
    },
    "input": {
        "pfs": {
            "glob": "/",
            "repo": "images"
        }
    },
    "transform": {
        "cmd": ["sh" ],
        "stdin": ["env"],
        "image": "ubuntu:14.04"
    },
    "enable_stats": true
}
```

Run this pipeline and upon completion you can view the log with
variables by running the following command:

```shell
pachctl logs --pipeline=env
PPS_WORKER_IP=172.17.0.7
DASH_PORT_8081_TCP_PROTO=tcp
PACHD_PORT_600_TCP_PORT=600
KUBERNETES_SERVICE_PORT=443
KUBERNETES_PORT=tcp://10.96.0.1:443
...
```

You should see a lengthy list of variables. Many of them define
internal networking parameters that most probably you will not
need to use.

Most users find the following environment variables
particularly useful:

| Environment Variable       | Description |
| -------------------------- | --------------------------------------------- |
| `PACH_JOB_ID`              | The ID of the current job. For example, <br> `PACH_JOB_ID=8991d6e811554b2a8eccaff10ebfb341`. |
| `PACH_OUTPUT_COMMIT_ID`    | The ID of the commit in the output repo for <br> the current job. For example, <br> `PACH_OUTPUT_COMMIT_ID=a974991ad44d4d37ba5cf33b9ff77394`. |
| `PPS_NAMESPACE`            | The PPS namespace. For example, <br> `PPS_NAMESPACE=default`. |
| `PPS_SPEC_COMMIT`          | The hash of the pipeline specification commit.<br> This value is tied to the pipeline version. Therefore, jobs that use <br> the same version of the same pipeline have the same spec commit. <br> For example, `PPS_SPEC_COMMIT=3596627865b24c4caea9565fcde29e7d`. |
| `PPS_POD_NAME`             | The name of the pipeline pod. For example, <br>`pipeline-env-v1-zbwm2`. |
| `PPS_PIPELINE_NAME`        | The name of the pipeline that this pod runs. <br> For example, `env`. |
| `PIPELINE_SERVICE_PORT_PROMETHEUS_METRICS` | The port that you can use to <br> exposed metrics to Prometheus from within your pipeline. The default value is 9090. |
| `HOME`                     | The path to the home directory. The default value is `/root` |
| `<input-repo>=<path/to/input/repo>` | The path to the filesystem that is <br> defined in the `input` in your pipeline specification. Pachyderm defines <br> such a variable for each input. The path is defined by the `glob` pattern in the <br> spec. For example, if you have an input `images` and a glob pattern of `/`, <br> Pachyderm defines the `images=/pfs/images` variable. If you <br> have a glob pattern of `/*`, Pachyderm matches <br> the files in the `images` repository and, therefore, the path is <br> `images=/pfs/images/liberty.png`. |
| `input_COMMIT`             | The ID of the commit that is used for the input. <br>For example, `images_COMMIT=fa765b5454e3475f902eadebf83eac34`. |
| `S3_ENDPOINT`         | A Pachyderm S3 gateway sidecar container endpoint. <br> If you have an S3 enabled pipeline, this parameter specifies a URL that <br> you can use to access the pipeline's repositories state when a <br> particular job was run. The URL has the following format: <br> `http://<job-ID>-s3:600`. <br> An example of accessing the data by using AWS CLI looks like this: <br>`echo foo_data | aws --endpoint=${S3_ENDPOINT} s3 cp - s3://out/foo_file`. |

In addition to these environment variables, Kubernetes injects others for
Services that run inside the cluster. These variables enable you to connect to
those outside services, which can be powerful but might also result
in processing being retried multiple times. For example, if your code
writes a row to a database, that row might be written multiple times because of
retries. Interaction with outside services must be idempotent to prevent
unexpected behavior. Furthermore, one of the running services that your code
can connect to is Pachyderm itself. This is generally not recommended as very
little of the Pachyderm API is idempotent, but in some specific cases it can be
a viable approach.

!!! note "See Also"
    - [transform.env](../../../reference/pipeline-spec/#transform-required)

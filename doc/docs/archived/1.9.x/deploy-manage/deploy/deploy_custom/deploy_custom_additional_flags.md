# Additional Flags

This section describes all the additional flags that you can use
to configure your custom deployment:

**Access to Kubernetes resources flags:**

* `--local-roles`: You can use the `--local-roles` flag to change
the kind of role the `pachyderm` service account uses
from cluster-wide (`ClusterRole`) to namespace-specific (`Role`).
Using `--local-roles` inhibits your ability to use the
[coefficient parallelism](http://docs.pachyderm.io/en/latest/reference/pipeline_spec.html#parallelism-spec-optional)
feature. After you set the `--local-roles` flag,
you might see a message similar to this in the `pachd` pod Kubernetes logs:

  ```
  ERROR unable to access kubernetes nodeslist, Pachyderm will continue to work 
  but it will not be possible to use COEFFICIENT parallelism. error: nodes is 
  forbidden: User "system:serviceaccount:pachyderm-test-1:pachyderm" cannot 
  list nodes at the cluster scope
  ```

**Resource requests and limits flags:**

Larger deployments might require you to configure more resources
for `pachd` and `etcd` or set higher limits for transient workloads.
The following flags set attributes which are passed on to
Kubernetes directly through the produced manifest.

* `--etcd-cpu-request`: The number of CPU cores that Kubernetes
allocates to `etcd`. Fractions are allowed.
* `--etcd-memory-request`: The amount of memory that Kubernetes
allocates to `etcd`. The SI suffixes are accepted as possible values.
* `--no-guaranteed`: Turn off QoS for `etcd` and `pachd`.
Do not use this flag in production environments.
* `--pachd-cpu-request`: The number of CPU cores that Kubernetes
allocates to `pachd`. Fractions are allowed.
* `--pachd-memory-request`: The amount of memory that Kubernetes
allocates to `pachd`. This flag accepts the SI suffixes.
* `--shards`: The maximum number of `pachd` nodes allowed in the
cluster. Increasing this number from the default value of `16`
might result in degraded performance.

**Note:** Do not modify the default values of these flags for
production deployments without consulting with Pachyderm support.

**Enterprise Edition flags:**

* `--dash-image`: The Docker image for the Pachyderm Enterprise Edition dashboard.
* `--image-pull-secret`: The name of a Kubernetes secret that Pachyderm uses to pull from a private Docker registry.
* `--no-dashboard`: Skip the creation of a manifest for the Enterprise Edition dashboard.
* `--registry`: The registry for Docker images.
* `--tls`:  A string in the `"<cert path>,<key path>"` format with the signed TLS certificate that is used for encrypting `pachd` communications.

**Output formats flags:**

* `--dry-run`: Create a manifest and send it to standard output, but do not deploy to Kubernetes.
* `-o` or `--output`: An output format. You can choose from JSON (default) or YAML.

**Logging flags:**

* `log-level`: The `pachd` verbosity level, from most verbose to least. You can set this parameter to `debug`, `info`, or `error`.
* `-v` or `--verbose`: Controls the verbosity of the `pachctl` invocation.


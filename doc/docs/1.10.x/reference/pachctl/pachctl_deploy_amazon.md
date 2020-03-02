## pachctl deploy amazon

Deploy a Pachyderm cluster running on AWS.

### Synopsis

Deploy a Pachyderm cluster running on AWS.
  <bucket-name>: An S3 bucket where Pachyderm will store PFS data.
  <region>: The AWS region where Pachyderm is being deployed (e.g. us-west-1)
  <disk-size>: Size of EBS volumes, in GB (assumed to all be the same).

```
pachctl deploy amazon <bucket-name> <region> <disk-size> [flags]
```

### Options

```
      --cloudfront-distribution string   Deploying on AWS with cloudfront is currently an alpha feature. No security restrictions have beenapplied to cloudfront, making all data public (obscured but not secured)
      --credentials string               Use the format "<id>,<secret>[,<token>]". You can get a token by running "aws sts get-session-token".
      --disable-ssl                      (rarely set) Disable SSL.
  -h, --help                             help for amazon
      --iam-role string                  Use the given IAM role for authorization, as opposed to using static credentials. The given role will be applied as the annotation iam.amazonaws.com/role, this used with a Kubernetes IAM role management system such as kube2iam allows you to give pachd credentials in a more secure way.
      --max-upload-parts int             (rarely set) Set a custom maximum number of upload parts. (default 10000)
      --no-verify-ssl                    (rarely set) Skip SSL certificate verification (typically used for enabling self-signed certificates).
      --part-size int                    (rarely set) Set a custom part size for object storage uploads. (default 5242880)
      --retries int                      (rarely set) Set a custom number of retries for object storage requests. (default 10)
      --reverse                          (rarely set) Reverse object storage paths. (default true)
      --timeout string                   (rarely set) Set a custom timeout for object storage requests. (default "5m")
      --upload-acl string                (rarely set) Set a custom upload ACL for object storage uploads. (default "bucket-owner-full-control")
      --vault string                     Use the format "<address/hostport>,<role>,<token>".
```

### Options inherited from parent commands

```
      --block-cache-size string         Size of pachd's in-memory cache for PFS files. Size is specified in bytes, with allowed SI suffixes (M, K, G, Mi, Ki, Gi, etc).
      --cluster-deployment-id string    Set an ID for the cluster deployment. Defaults to a random value.
  -c, --context string                  Name of the context to add to the pachyderm config. If unspecified, a context name will automatically be derived.
      --create-context --dry-run        Create a context, even with --dry-run.
      --dash-image string               Image URL for pachyderm dashboard
      --dashboard-only                  Only deploy the Pachyderm UI (experimental), without the rest of pachyderm. This is for launching the UI adjacent to an existing Pachyderm cluster. After deployment, run "pachctl port-forward" to connect
      --dry-run --create-context        Don't actually deploy pachyderm to Kubernetes, instead just print the manifest. Note that a pachyderm context will not be created, unless you also use --create-context.
      --dynamic-etcd-nodes int          Deploy etcd as a StatefulSet with the given number of pods.  The persistent volumes used by these pods are provisioned dynamically.  Note that StatefulSet is currently a beta kubernetes feature, which might be unavailable in older versions of kubernetes.
      --etcd-cpu-request string         (rarely set) The size of etcd's CPU request, which we give to Kubernetes. Size is in cores (with partial cores allowed and encouraged).
      --etcd-memory-request string      (rarely set) The size of etcd's memory request. Size is in bytes, with SI suffixes (M, K, G, Mi, Ki, Gi, etc).
      --etcd-storage-class string       If set, the name of an existing StorageClass to use for etcd storage. Ignored if --static-etcd-volume is set.
      --expose-object-api               If set, instruct pachd to serve its object/block API on its public port (not safe with auth enabled, do not set in production).
      --image-pull-secret string        A secret in Kubernetes that's needed to pull from your private registry.
      --local-roles                     Use namespace-local roles instead of cluster roles. Ignored if --no-rbac is set.
      --log-level string                The level of log messages to print options are, from least to most verbose: "error", "info", "debug". (default "info")
      --namespace string                Kubernetes namespace to deploy Pachyderm to.
      --new-storage-layer               (feature flag) Do not set, used for testing.
      --no-color                        Turn off colors.
      --no-dashboard                    Don't deploy the Pachyderm UI alongside Pachyderm (experimental).
      --no-expose-docker-socket         Don't expose the Docker socket to worker containers. This limits the privileges of workers which prevents them from automatically setting the container's working dir and user.
      --no-guaranteed                   Don't use guaranteed QoS for etcd and pachd deployments. Turning this on (turning guaranteed QoS off) can lead to more stable local clusters (such as on Minikube), it should normally be used for production clusters.
      --no-rbac                         Don't deploy RBAC roles for Pachyderm. (for k8s versions prior to 1.8)
  -o, --output string                   Output format. One of: json|yaml (default "json")
      --pachd-cpu-request string        (rarely set) The size of Pachd's CPU request, which we give to Kubernetes. Size is in cores (with partial cores allowed and encouraged).
      --pachd-memory-request string     (rarely set) The size of PachD's memory request in addition to its block cache (set via --block-cache-size). Size is in bytes, with SI suffixes (M, K, G, Mi, Ki, Gi, etc).
      --registry string                 The registry to pull images from.
      --require-critical-servers-only   Only require the critical Pachd servers to startup and run without errors.
      --shards int                      (rarely set) The maximum number of pachd nodes allowed in the cluster; increasing this number blindly can result in degraded performance. (default 16)
      --static-etcd-volume string       Deploy etcd as a ReplicationController with one pod.  The pod uses the given persistent volume.
      --tls string                      string of the form "<cert path>,<key path>" of the signed TLS certificate and private key that Pachd should use for TLS authentication (enables TLS-encrypted communication with Pachd)
      --upload-concurrency-limit int    The maximum number of concurrent object storage uploads per Pachd instance. (default 100)
  -v, --verbose                         Output verbose logs
```


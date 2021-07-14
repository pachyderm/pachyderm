# Config Specification

This document outlines the fields in pachyderm configs. This should act as a
reference. If you wish to change a config value, you should do so via
`pachctl config`.

## JSON format

```json
{
  "user_id": string,
  "v2": {
    "active_context": string,
    "active_enterprise_context": string,
    "contexts": {
        "source": int,
        "pachd_address": string,
        "server_cas": string,
        "session_token": string,
        "active_transaction": string,
        "cluster_name": string,
        "auth_info": string,
        "namspace": string,
        "cluster_deployment_id": string,
        "enterprise_server": bool,
        "port_forwarders": {
          service_name: int,
          ...
        }
      ...
    },
    "metrics": bool
  }
}
```

If a field is not set, it will be omitted from JSON entirely. Following is an
example of a simple config:

```json
{
  "user_id": "514cbe16-e615-46fe-92d9-3156f12885d7",
  "v2": {
    "active_context": "default",
    "contexts": {
      "default": {}
    },
    "metrics": true
  }
}
```

Following is a walk-through of all the fields.

### User ID

A UUID giving a unique ID for this user for metrics.

### Metrics

Whether metrics is enabled.

### Active Context

`v2.active_context` specifies the name of the currently active pachyderm
context, as specified in `v2.contexts`.

### Active Enterprise Context

`v2.active_enterprise_context` specifies the name of the currently active pachyderm enterprise context, as specified in `v2.contexts`. If left blank the `v2.active_context` value will be interpreted as the Active Enterprise Context.

### Contexts

A map of context names to their configurations. Pachyderm contexts are akin to
kubernetes contexts (and in fact reference the kubernetes context that they're
associated with.)

#### Source

An integer that specifies where the config came from. This parameter is for internal use only and
should not be modified.

#### Pachd Address

A `host:port` specification for connecting to pachd. If this is set, pachyderm
will directly connect to the cluster, rather than resorting to kubernetes'
port forwarding. If you can set this (because there's no firewall between you
and the cluster), you should, as kubernetes' port forwarder is not designed to
handle large amounts of data.

#### Server CAs

Trusted root certificates for the cluster, formatted as a base64-encoded PEM.
This is only set when TLS is enabled.

#### Session token

A secret token identifying the current pachctl user within their pachyderm
cluster. This is included in all RPCs sent by pachctl, and used to determine
if pachctl actions are authorized. This is only set when auth is enabled.

#### Active transaction

The currently active transaction for batching together pachctl commands. This
can be set or cleared via many of the `pachctl * transaction` commands.

#### Cluster name

The name of the underlying Kubernetes cluster, extracted from the Kubernetes
context.

#### Auth info

The name of the underlying Kubernetes cluster's auth credentials, extracted
from the Kubernetes context.

#### Namespace

The underlying Kubernetes cluster's namespace, extracted from the Kubernetes
context.

#### Cluster Deployment ID

The pachyderm cluster deployment ID that is used to ensure the operations run on the
expected cluster.

#### Enterprise Server
Whether the context represents an enterprise server.

#### Port forwarders

A mapping of `service name -> local port`. This field is populated when you run
explicit port forwarding (`pachctl port-forward`), so that subsequent
`pachctl` operations know to use the explicit port forwarder.

This field is removed when the `pachctl port-forward` operation
completes. You might need to manually delete the field from your
config if the process failed to remove the field automatically.

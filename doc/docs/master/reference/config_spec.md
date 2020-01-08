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
    "contexts": {
      string: {
        "source": int,
        "pachd_address": string,
        "server_cas": string,
        "session_token": string,
        "active_transaction": string,
        "cluster_name": string,
        "auth_info": string,
        "namspace": string,
        "cluster_id": string,
        "port_forwarders": {
          service_name: int,
          ...
        }
      },
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

`v2.active_context` specifies the name of the currently actively pachyderm
context, as specified in `v2.contexts`.

### Contexts

A map of context names to their configurations. Pachyderm contexts are akin to
kubernetes contexts (and in fact reference the kubernetes context that they're
associated with.)

#### Source

An int specifying where the config came from. This is for internal use, and
shouldn't be modified.

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

The name of the underlying kubernetes cluster, extracted from the kubernetes
context.

#### Auth info

The name of the underlying kubernetes cluster's auth credentials, extracted
from the kubernetes context.

#### Namespace

The underlying kubernetes cluster's namespace, extracted from the kubernetes
context.

#### Cluster ID

The pachyderm cluster ID - used to ensure we're running operations on the
expected cluster.

#### Port forwarders

A mapping of service name -> local port. This field is populated when running
explicit port forwarding (`pachctl port-forward`), so that subsequent
`pachctl` operations know to use the explicit port forwarder.

This field will be removed when the `pachctl port-forward` operation
completes, but note that you may need to manually delete the field from your
config if the process was killed being being able to do so.

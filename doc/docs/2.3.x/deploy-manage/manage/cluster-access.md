# Manage Cluster Access

Pachyderm contexts enable you to store configuration parameters for
multiple Pachyderm clusters in a single configuration file saved at
`~/.pachyderm/config.json`. This file stores the information
about all Pachyderm clusters that you have deployed from your
machine locally or on a remote server.

For example, if you have a cluster that
is deployed locally in `minikube` and another one deployed on
Amazon EKS, configurations for these clusters are stored in
that `config.json` file. By default, all local cluster configurations
have the `local` prefix. If you have multiple local clusters,
Pachyderm adds a consecutive number to the `local` prefix
of each cluster.

The following text is an example of a Pachyderm `config.json` file:

```shell
{
   "user_id": "b4fe4317-be21-4836-824f-6661c68b8fba",
   "v2": {
     "active_context": "local-1",
     "contexts": {
       "default": {},
       "local": {},
       "local-1": {},
     },
     "metrics": true
   }
}
```

## View the Active Context

When you have multiple Pachyderm clusters, you can switch
between them by setting the current context.
The active context is the cluster that you interact with when
you run `pachctl` commands.

To view active context, type:

* View the active context:

    ```shell
    pachctl config get active-context
    ```

    **System response:**

    ```shell
    local-1
    ```

* List all contexts and view the current context:

    ```shell
    pachctl config list context
    ```

    **System response:**

    ```shell
      ACTIVE  NAME
              default
              local
      *       local-1
    ```

    The active context is marked with an asterisk.

## Change the Active Context

To change the active context, type `pachctl config set
active-context <name>`.

Also, you can set the `PACH_CONTEXT` environmental variable
that overrides the active context.

**Example:**

```shell
export PACH_CONTEXT=local1
```

## Create a New Context

When you deploy a new Pachyderm cluster, a new context
that points to the new cluster is created automatically.

In addition, you can create a new context by providing your parameters
through the standard input stream (`stdin`) in your terminal.
Specify the parameters as a comma-separated list enclosed in
curly brackets.

!!! note
    By default, the `pachd` port is `30650`.

To create a new context with specific parameters, complete
the following steps:

1. Create a new Pachyderm context with a specific `pachd` IP address
and a client certificate:

    ```shell
    echo '{"pachd_address":"10.10.10.130:650", "server_cas":"insert your base 64 encoded key.pem"}' | pachctl config set context new-local
    ```

    **System response:**

    ```shell
    Reading from stdin
    ```

1. Verify your configuration by running the following command:

    ```shell
    pachctl config get context new-local
    {
      "pachd_address": "10.10.10.130:650",
      "server_cas": "insert your base 64 encoded key.pem"
    }
    ```

## Update an Existing Context

You can update an existing context with new parameters, such
as a Pachyderm IP address, certificate authority (CA), and others.
For the list of parameters, see [Pachyderm Config Specification](../../reference/config-spec.md).

To update the Active Context, run the following commands:

1. Update the context with a new `pachd` address:

    ```shell
    pachctl config update context local-1 --pachd-address 10.10.10.131
    ```

    The `pachctl config update` command supports the `--pachd-address`
    flag only.

1. Verify that the context has been updated:

    ```shell
    pachctl config get context local-1
    ```

    **System response:**

    ```shell
    {
      "pachd_address": "10.10.10.131"
    }
    ```

1. Alternatively, you can update multiple properties by using
an `echo` script:

    ```shell
    echo '{"pachd_address":"10.10.10.132", "server_cas":"insert your base 64 encoded key.pem"}' | pachctl config set context local-1 --overwrite
    ```

    **System response:**

    ```shell
    Reading from stdin.
    ```

1. Verify that the changes were applied:

    ```shell
    pachctl config get context local-1
    ```

    **System response:**

    ```shell
    {
      "pachd_address": "10.10.10.132",
      "server_cas": "LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCk1JSUVEakNDQXZhZ0F3SUJBZ0lERDkyc01BMEdDU3FHU0liM0RRRUJDd1VBTUVVeEN6QUpCZ05WQkFZVEFrUkYKTVJVd0V3WURWUVFLREF4RUxWUnlkWE4wSUVkdFlrZ3hIekFkQmdOVkJBTU1Ga1F0VkZKVlUxUWdVbTl2ZENCRApRU0F6SURJd01UTXdIaGNOTVRNd09USXdNRGd5TlRVeFdoY05Namd3T1RJd01EZ3lOVFV4V2pCRk1Rc3dDUVlEClZRUUdFd0pFUlRFVk1CTUdBMVVFQ2d3TVJDMVVjblZ6ZENCSGJXSklNUjh3SFFZRFZRUUREQlpFTFZSU1ZWTlUKSUZKdmIzUWdRMEVnTXlBeU1ERXpNSUlCSWpBTkJna3Foa2lHOXcwQkFRRUZBQU9DQVE4QU1JSUJDZ0tDQVFFQQp4SHRDa29JZjdPMVVtSTRTd01vSjM1TnVPcE5jRytRUWQ1NU9hWWhzOXVGcDh2YWJvbUd4dlFjZ2RKaGw4WXdtCkNNMm9OY3FBTnRGamJlaEVlb0xEYkY3ZXUrZzIwc1JvTm95Zk1yMkVJdURjd3U0UVJqbHRyNU01cm9mbXc3d0oKeVN4cloxdlptM1oxVEF2Z3U4WFh2RDU1OGwrKzBaQlgrYTcyWmw4eHY5TnRqNmU2U3ZNalpidTM3Nk1sMXdycQpXTGJ2aVByNmViSlNXTlh3ckl5aFVYUXBsYXBSTzVBeUE1OGNjblNRM2ozdFlkTGw0LzFrUitXNXQwcXA5eCt1CmxvWUVyQy9qcElGM3Qxb1cvOWdQUC9hM2VNeWtyL3BiUEJKYnFGS0pjdStJODlWRWdZYVZJNTk3M2J6Wk5POTgKbER5cXdFSEM0NTFRR3NEa0dTTDhzd0lEQVFBQm80SUJCVENDQVFFd0R3WURWUjBUQVFIL0JBVXdBd0VCL3pBZApCZ05WSFE0RUZnUVVQNURJZmNjVmIvTWtqNm5ETDB1aUR5R3lMK2N3RGdZRFZSMFBBUUgvQkFRREFnRUdNSUcrCkJnTlZIUjhFZ2JZd2diTXdkS0J5b0hDR2JteGtZWEE2THk5a2FYSmxZM1J2Y25rdVpDMTBjblZ6ZEM1dVpYUXYKUTA0OVJDMVVVbFZUVkNVeU1GSnZiM1FsTWpCRFFTVXlNRE1sTWpBeU1ERXpMRTg5UkMxVWNuVnpkQ1V5TUVkdApZa2dzUXoxRVJUOWpaWEowYVdacFkyRjBaWEpsZG05allYUnBiMjVzYVhOME1EdWdPYUEzaGpWb2RIUndPaTh2ClkzSnNMbVF0ZEhKMWMzUXVibVYwTDJOeWJDOWtMWFJ5ZFhOMFgzSnZiM1JmWTJGZk0xOHlNREV6TG1OeWJEQU4KQmdrcWhraUc5dzBCQVFzRkFBT0NBUUVBRGxrT1dPUjBTQ05FenpRaHRad1VHcTJhUzdlemlHMWNxUmR3OENxZgpqWHY1ZTRYNnh6bm9FQWl3TlN0Znp3TFMwNXpJQ3g3dUJWU3VONU1FQ1gxc2o4SjB2UGdjbEw0eEFVQXQ4eVFnCnQ0UlZMRnpJOVhSS0VCbUxvOGZ0TmRZSlNOTU93TG81cUxCR0FyRGJ4b2had3I3OGU3RXJ6MzVpaDFXV3pBRnYKbTJjaGxUV0wrQkQ4Y1J1M1N6ZHBwanZXN0l2dXdiRHpKY21Qa24yaDZzUEtSTDhtcFhTU25PTjA2NTEwMmN0TgpoOWo4dEdsc2k2QkRCMkI0bCtuWmszekNScnliTjFLajdZbzhFNmw3VTB0Sm1oRUZMQXR1VnF3ZkxvSnM0R2xuCnRRNXRMZG5rd0JYeFAvb1ljdUVWYlNkYkxUQW9LNTlJbW1Rcm1lL3lkVWxmWEE9PQotLS0tLUVORCBDRVJUSUZJQ0FURS0tLS0tCg=="
    }
    ```

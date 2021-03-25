# Deploy Pachyderm with TLS

You can deploy your Pachyderm cluster with Transport Layer Security(TLS)
enabled to ensure your cluster communications are protected from external
attackers, and all the communication parties are verified by means of a
trusted certificate and a private key. For many organizations, TLS is a
security requirement that ensures integrity of their data.
Before you can enable TLS, you need to obtain a certificate from a trusted
CA, such as Let's Encrypt, Cloudflare, or other.
You can enable TLS during the deployment of your Pachyderm cluster by
providing a path to your CA certificate and your private key by using the
`--tls` flag with the `pachctl deploy` command.

```shell
$ pachctl deploy <platform> --tls "<path/to/cert>,<path/to/key>"
```

!!! note
    The paths to the certificate and to the key must be specified
    exactly as shown in the example above — in double quotes, separated by
    a comma, and without a space.

After you deploy Pachyderm, to connect through `pachctl` by using a
trusted certificate, you need to configure the `pachd_address` in the
Pachyderm context with the cluster IP address that starts with `grpcs://`.
You can do so by running the following command:

!!! example
    ```shell
    echo '{"pachd_address": "grpcs://<cluster-ip>:31400"}' | pachctl config
    pachctl config update context `p config get active-context` --pachd_address "grpcs://<cluster-ip>:31400"
    ```

!!! note "See Also:"

- [Connect by using a Pachyderm context](../connect-to-cluster/#connect-by-using-a-pachyderm-context)

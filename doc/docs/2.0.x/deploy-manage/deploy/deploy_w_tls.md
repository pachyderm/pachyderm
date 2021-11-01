# Deploy Pachyderm with TLS

You can deploy your Pachyderm cluster with Transport Layer Security (TLS)
enabled to ensure your cluster communications are protected from external
attackers, and all the communication parties are verified by means of a
trusted certificate and a private key. For many organizations, TLS is a
security requirement that ensures integrity of their data.
Before you can enable TLS, you need to obtain a certificate from a trusted
CA, such as Let's Encrypt, Cloudflare, or other.
You can enable TLS during the deployment of your Pachyderm cluster by configuring it in your helm values. You can either provide an existing secret with your certificate, or supply your certificate and key and helm will create the secret for you.

```yaml
pachd:
  tls:
    enabled: true
    secretName: ""
    newSecret:
      create: false
      crt: ""
      key: ""
```

After you deploy Pachyderm, to connect through `pachctl` by using a
trusted certificate, you need to configure the `pachd_address` in the
Pachyderm context with the cluster IP address that starts with `grpcs://`.
You can do so by running the following command:

!!! example
    ```shell   
    echo '{"pachd_address": "grpcs://<cluster-ip:30650"}' | pachctl config set context "local-grpcs" --overwrite && pachctl config set active-context "local-grpcs"   
    ```

!!! note "See Also:"

- [Connect by using a Pachyderm context](../connect-to-cluster/#connect-by-using-a-pachyderm-context)

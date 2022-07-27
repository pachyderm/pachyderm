# Deploy Pachyderm with TLS (SSL, HTTPS)

## Obtain A Certificate And Enable TLS

You can deploy your Pachyderm cluster with Transport Layer Security (TLS)
enabled to secure internet browser connections and transactions through data encryption by means of a trusted certificate and a private key. 

Before you can enable TLS:

- Obtain a certificate from a trusted Certificate Authority such as [Let's Encrypt](https://letsencrypt.org/){target=_blank}, [HashiCorp Vault](https://www.vaultproject.io/){target=_blank}, [Venafi](https://www.venafi.com/){target=_blank}... 
- Create a tls secret ( `kubectl create secret tls <name> --key=tls.key --cert=tls.cert`) with the  "tls.key" and "tls.crt" keys containing the PEM-encoded private key and certificate material.

Optionally, you can install [Cert-Manager](https://cert-manager.io/docs/installation/){target=_blank} on your cluster to simplify the process of obtaining (No Certificate Signing Requests needed), renewing, and using certificates. 
In particular, you can use cert-manager to:

- **Talk to a certificate issuer**  Cert-manager comes with a number of built-in certificate issuers. You can also install external issuers in addition to the built-in types.

- **Obtain your certificate**:

    You can verify that the certificate is issued correctly by running the following command:

    ```shell
    kubectl get certificate
    ```
    You should see the certificate with a status of Ready in output.

- **Create the backing tls secret** holding your Certificate and private key automatically.

Once your tls secret is created:

- Enable tls in your helm values.
- Reference this certificate object in your helm chart by setting your tls secret name in the proper tls section. (For the Cert Manager users, the secret name should match the name set in your [certificate ressource](https://cert-manager.io/docs/usage/certificate/#creating-certificate-resources){target=_blank}.

!!! Example
    In this example, you terminate tls at the cluster level by enabling tls directly on pachd:
    
    ```yaml
     pachd:
       tls:
          enabled: true
          secretName: "<the-secret-name-in-your-certificate-ressource>"
    ```

Et voila!

!!! Note
    When using self signed certificates or custom certificate authority, you will need to set `global.customCaCerts` to true to add Pachyderm's certificate and CA to the list of trusted authorities for console and enterprise, allowing Pachyderm components (pachd, Console, Enterprise Server) to communicate over SSL. 

    If you are using a custom ca-signed cert, **you must include the full certificate chain in the root.crt file**.


!!! Attention 
    We are now shipping Pachyderm with an **embedded proxy** 
    allowing your cluster to expose one single port externally. 
    This deployment setup is optional.
    
    If you choose to deploy Pachyderm with a proxy (see the [deployment instructions](../deploy-w-proxy/) and new recommended architecture), the setup of **tls is set in the proxy section of your values.yaml** only (i.e., tls terminates inside the proxy).

    The setup of TLS at the proxy level is intended for the case where the proxy is exposed directly to the Internet.

    ```yaml
      proxy:
        tls:
          enabled: true
          secretName: "<the-secret-name-in-your-certificate-ressource>"
    ```

## Connect to Pachyderm Via SSL

After you deploy Pachyderm, to connect through `pachctl` by using a
trusted certificate, you will need to set the `pachd_address` in the
Pachyderm context with the cluster IP address that starts with `grpcs://`.
You can do so by running the following command:

!!! example
    ```shell   
    echo '{"pachd_address": "grpcs://<cluster-ip:30650"}' | pachctl config set context "grpcs-context" --overwrite && pachctl config set active-context "grpcs-context"   
    ```

!!! Attention "Attention proxy users, your port number is now 443"

    ```shell
    echo '{"pachd_address": "grpcs://<external-IP-address-or-domain-name>:443"}' | pachctl config set context "grpcs-context" --overwrite
    ```
    ```shell
    pachctl config set active-context "grpcs-context"
    ```

!!! note "See Also:"
    [Connect by using a Pachyderm context](../connect-to-cluster/#connect-by-using-a-pachyderm-context)

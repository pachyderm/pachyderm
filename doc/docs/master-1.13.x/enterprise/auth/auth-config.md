# Manage Authentication Configuration

The Pachyderm authentication configuration file includes 
information about the identity provider (IdP) configured in Pachyderm.
The file is stored in the Pachyderm etcd container, and you
can access it by using the following commands:

* To view the auth configuration, run:

  ```shell
  pachctl auth get-config
  ```

* To edit the auth configuration, run:

  ```shell
  pachctl auth set-config -f <config.json>
  ```

  or:

  ```shell
  pachctl auth set-config <<EOF
  {
    "live_config_version": ${live_config_version},

    "id_providers": [
      {
       ...
      }
    ]
    ...
  }
  EOF
  ```

* To delete Pachyderm auth configuration, run:

  ```shell
  pachctl auth deactivate
  ```

## SAML Authentication Parameters

You can specify the following parameters for your SAML provider in the
authentication file:

| Parameter        | Description                          |
| ---------------- | ------------------------------------ |
| `name`           | The name of the SAML provider. For example, `auth0`. <br> This name is used as a prefix for all usernames derived <br> from the identity  provider. For example, <br> `auth0:test@pachyderm.com`. |
| `description`    | An optional description of the identity provider. |
| `saml`           | A list of parameters related to the SAML provider <br> configuration. |
| `metadata_url`   | A URL of the SAML provider metadata service. |
| `metada_xml`     | The XML metadata of SAML IdP. You can use this <br> parameter if the IdP is located in another network to which <br> users have access, but `pachd` does not. It can <br> also be used for testing when the IdP is not yet <br> configured. |
| `group_attribute` | A group configured on the IdP. The parameters enable <br> you to grant permissions on at a group level rather <br> than on an individual level. |
| `saml_svc_options` | A list of options for SAML services |
| `acs_url`          | The URL of the `pachd`'s Assertion Consumer Service <br> and Metadata Service (ACS). If Pachyderm runs in a <br> private cluster, the cluster admin must set up <br> the domain name and proxy to resolve to <br> `pachd:654/acs`. For example, <br> `http://localhost:30654/saml/acs`. |
| `metadata_url`    | The public URL of Pachd's SAML metadata service. <br> This parameter under the `saml_svc_options` is <br> different from the one under the `saml` option. <br> If Pachyderm runs in a private cluster, you must <br> create this URL, which resolves to <br> `pachd:654/saml/metadata`. For example, <br>`http://localhost:30654/saml/metadata`. |
| `dash_url`         | The public URL of the Pachyderm dashboard. <br> For example, `https://localhost:30080`. |
| `session_duration` | The length of a user session in hours (h) or <br> minutes (m). For example, `8h`. If left blank 24 hours session is <br> configured by default. |

[View a sample config](../saml/#write-pachyderm-config)

## OIDC Authentication Parameters

 You can specify the following parameters for your OIDC provider in the
 authentication file:

 | Parameter        | Description                          |
 | ---------------- | ------------------------------------ |
 | `name`           | The name of the OIDC provider. For example, <br> `keycloak`. This name is used as a prefix for all usernames derived <br> from the identity provider. For example, <br> `keycloak:test@pachyderm.com`. |
 | `description`    | An optional description of the identity provider. |
 | `oidc`           | A list of parameters related to the OIDC provider configuration. |
 | `issuer`         | The address of the OIDC provider. For example, <br> `http://keycloak.<ip>.nip.io/auth/realms/<realm-name>`. | 
 | `client_id`      | The Pachyderm ID configured in the IdP. For example, <br> `pachyderm`.
 | `client_secret`  | A shared secret with the ID provider. If your OIDC provider <br> does not use a secret, which is not recommended, the <br> parameter can be omitted for testing. |
 | `redirect_uri`   | The URI on which the OIDC IdP can access Pachyderm. <br> Depends on your network configuration and must have the following <br> format: `http://<ip>:30657/authorization-code/callback`. |
 | `additional_scopes`| A list of additional OIDC scopes to request from the provider. If `groups` is requested, the groups in the ID token will be synced with Pachyderm |


[View a sample config](../oidc/configure-keycloak/#configure-keycloak)

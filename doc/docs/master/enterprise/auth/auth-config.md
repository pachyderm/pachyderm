# Manage Authentication Configuration

The Pachyderm authentication configuration files includes 
information about the identity provider configured in Pachyderm.
The file is stored in the Pachyderm etcd container and you
can access it by using the following commands:

* To view the auth configuration, run:

  ```bash
  pachctl auth get-config
  ```

* To edit the auth configuration, run:

  ```bash
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

  ```bash
  pachctl auth deactivate
  ```

## SAML Authentication Parameters

You can specify the following parameters for your SAML provider in the
authentication file:

| Parameter        | Description                          |
| ---------------- | ------------------------------------ |
| `name`           | The name of the SAML provider. For example, `okta`. <br> This name is used as a prefix for all usernames derived from the identity provider. For example, `okta:test@pachyderm.com`. |
| `description`    | An optional description of the identity provider. |
| `saml`           | A list of parameters related to the SAML provider configuration. |
| `metadata_url`   | A URL of the SAML provider metadata service. |
| `metada_xml`     | The XML metadata of SAML IdP. You can use this parameter if the IdP is located in another network to which users have access, but `pachd` does not. It can also be used for testing when the IdP is not yet
configured. |
| `group_attribute` | A group configured on the IdP. The parameters enable you to grant permissions on at a group level rather than on an individual level. |
| `saml_svc_options` | A list of options for SAML services |
| `acs_url`          | The URL of the `pachd`'s Assertion Consumer Service and Metadata Service (ACS). If Pachyderm runs in a private cluster, the cluster admin must set up the domain name and proxy to resolve to `pachd:654/acs`. For example, `http://localhost:30654/saml/acs`. |
| `meatadata_url`    | The public URL of Pachd's SAML metadata service. This parameter under the `saml_svc_options` is different from the one under the `saml` option. If Pachyderm runs in a private cluster, you must create this URL, which must resolve to `pachd:654/saml/metadata`. For example, `http://localhost:30654/saml/metadata`. |
| `dash_url`         | The public URL of the Pachyderm dashboard. For example, `https://localhost:30080`. |
| `session_duration` | The length of a user session in hours (h) or minutes (m). For example, `8h`. If left blank 24 hours session is configured by default. |

[View a sample config](../saml/saml_setup/#write-pachyderm-config)

## OIDC Authentication Parameters

 You can specify the following parameters for your OIDC provider in the
 authentication file:

 | Parameter        | Description                          |
 | ---------------- | ------------------------------------ |
 | `name`           | The name of the OIDC provider. For example, `keycloak`. This name is used as a prefix for all usernames derived from the identity provider. For example, `keycloak:test@pachyderm.com`. |
 | `description`    | An optional description of the identity provider. |
 | `oidc`           | A list of parameters related to the OIDC provider configuration. |
 | `issuer`         | The address of the OIDC provider. For example, `http://keycloak.<ip>.nip.io/auth/realms/<realm-name>`. | 
 | `client_id`      | The Pachyderm ID configured in the IdP. For example, `pachyderm`.
 | `client_secret`  | A shared secret with the ID provider. This parameter can be omitted for testing. |
 | `redirect_uri`   | The URI on which the OIDC IdP can access Pachyderm. Depends on your network configuration and must have the following format: `http://<ip>:30657/authorization-code/callback`. |

[View a sample config](../oidc/configure-keycloak/#configure-keycloak)

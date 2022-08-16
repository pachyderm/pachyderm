# Configure Enterprise Variables 

## Variables

| Variable | Required | Type | Default | Description |
|---|---|---|---|---|
| ACTIVATE_AUTH | No | boolean | FALSE | Enables [role-based access to Pachyderm](../../../enterprise/auth/authorization/index.md).  |
| AUTH_ROOT_TOKEN | No | string |  | Contains the secret name used to pass the `rootToken` value via an existing Kubernetes secret. |
| AUTH_CONFIG | No | string |  | Contains the auth configuration; can be set with `pachctl auth set-config`. |
| AUTH_CLIENT_SECRET | No | string |  | Contains the client secret name. |
| AUTH_CLUSTER_RBAC | No | string |  | Contains the secret name for cluster role-based access settings.  |
| LICENSE_KEY | No | string |  | Contains the secret name for the license required to use enterprise Pachyderm; can be set with `pachctl license activate` |
| ENTERPRISE_SECRET | No | string |  | Contains the name of the secret used to pass an enterprise secret value via an existing Kubernetes secret. |
| ENTERPRISE_MEMBER | No | boolean | FALSE | Enables a user to use enterprise Pachyderm. |
| ENTERPRISE_SERVER_ADDRESS | No | string |  | Contains the server address for enterprise Pachyderm.  |
| ENTERPRISE_SERVER_TOKEN | No | string |  | Contains the secret name used to pass the `enterpriseServerToken` value via an existing Kubernetes secret. |
| ENTERPRISE_MEMBER_CONFIG | No | string |  | Contains the member configuration settings. |
| IDP_CONFIG | No | string |  | Contains the Pachyderm identity configuration settings; can be set with `pachctl idp set-config`. |
| IDP_CONNECTORS | No | string |  | Contains the secret name for IDP connectors. |
| IDP_CLIENTS | No | string |  | Contains the IDP clients list; can be defined with `pachctl idp create-client` and listed with `pachctl idp list-client`. |
| IDP_ADDITIONAL_CLIENTS | No | string |  | The list of additional clients for the cluster to recognize. |
| TRUSTED_PEERS | No | string |  | The list of identity services to recognize additional OIDC clients as trusted peers of pachd. |
| CONSOLE_OAUTH_ID | No | string |  | The Oauth ID for console. |
| CONSOLE_OAUTH_SECRET | No | string |  | The name of the secret used to pass the `OAUTH_CLIENT_SECRET` value via an existing Kubenetes secret. |
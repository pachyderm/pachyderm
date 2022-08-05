# Configure Enterprise Variables 

## Variables

| Variable | Required | Type | Default | Description |
|---|---|---|---|---|
| ACTIVATE_AUTH | No | boolean | FALSE |  |
| AUTH_ROOT_TOKEN | No | string |  | The name of the secret used to pass the `rootToken` value via an existing Kubernetes secret. |
| AUTH_CONFIG | No | string |  |  |
| AUTH_CLIENT_SECRET | No | string |  |  |
| AUTH_CLUSTER_RBAC | No | string |  |  |
| LICENSE_KEY | No | string |  | The license key required to use enterprise Pachyderm. |
| ENTERPRISE_SECRET | No | string |  | The name of the secret used to pass the enterprise secret value via an existing Kubernetes secret. |
| ENTERPRISE_MEMBER | No | boolean | FALSE | The flag for whether or not a user has enterprise Pachyderm. |
| ENTERPRISE_SERVER_ADDRESS | No | string |  |  |
| ENTERPRISE_SERVER_TOKEN | No | string |  | The name of the secret used to pass the `enterpriseServerToken` value via an existing Kubernetes secret. |
| ENTERPRISE_MEMBER_CONFIG | No | string |  |  |
| IDP_CONFIG | No | string |  |  |
| IDP_CONNECTORS | No | string |  |  |
| IDP_CLIENTS | No | string |  |  |
| IDP_ADDITIONAL_CLIENTS | No | string |  | The list of clients for the cluster to recognize. |
| TRUSTED_PEERS | No | string |  | The list of identity services to recognize additional OIDC clients as trusted peers of pachd. |
| CONSOLE_OAUTH_ID | No | string |  | The Oauth ID for console. |
| CONSOLE_OAUTH_SECRET | No | string |  | The name of the secret used to pass the `OAUTH_CLIENT_SECRET` value via an existing Kubenetes secret. |
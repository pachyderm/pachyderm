# Configure an OIDC Auth Provider

OpenID Connect (OIDC) is an identity layer that extends authorization
capabilities of the OAuth 2.0 security framework with authentication
and single sign-on (SSO). Pachyderm supports authentication providers,
such as [Keycloak](https://www.keycloak.org) and [Okta](https://www.okta.com/)
that work with OIDC.

Pachyderm designates port `657` for OIDC connections.
The redirect URL that an OIDC provider uses to forward connections
back to Pachyderm is `http://<ip>:657/authorization-code/callback`.

Your port number should be whatever is routing to pachd:657.

!!! note "See Also"
    - [Manage Authentication Configuration](../auth-config/) 

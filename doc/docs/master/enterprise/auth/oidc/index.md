# Configure an OIDC Auth Provider

OpenID Connect (OIDC) is an identity layer that extends authorization
capabilities of the OAuth 2.0 security framework with authentication
and single sign-on (SSO). Pachyderm supports authentication providers,
such as [Keycloak](https://www.keycloak.org) and [Okta](https://www.okta.com/)
that work with OIDC.

Pachyderm designates port `30657` for OIDC connections. This port can
be modified as needed. In a default testing Pachyderm deployment, the
root URL that needs to be used to access an OIDC provider is
`http://<ip>:30657/authorization-code/callback`.

!!! important
    Although other identity providers might work with Pachyderm, only
    Keycloak and Okta has been tested and is officially supported.

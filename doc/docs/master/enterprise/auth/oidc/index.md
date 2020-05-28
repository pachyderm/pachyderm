# Configure an OIDC Auth Provider

OpenID Connect (OIDC) is an identity layer that extends authorization
capabilities of the OAuth 2.0 security framework with authentication
and single sign-on (SSO). Pachyderm supports authentication providers,
such as [Keycloak](https://www.keycloak.org), that work with OIDC.

Pachyderm designates port `14687` for OIDC connections. This port can
be modified as needed. In a default testing Pachyderm deployment, the
root URL that needs to be used to access an OIDC provider is
`http://localhost:14687/authorization-code/callback`.


!!! important
    Although other identity providers might work with Pachyderm, only
    Keycloak has been tested and is officially supported. Keycloak's
    SAML support has not been tested.

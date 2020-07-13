# Configure Access Controls

!!! note
    Access Controls is an enterprise feature that requires
    an active enterprise token.

Pachyderm Access Controls enable you to log in to Pachyderm
with a user configured in a third-party identity management
platform and operate as that user with the data stored in
Pachyderm. Pachyderm supports the following identity providers
with the specified authentication protocols:

- Okta™ with Security Assertion Markup Language(SAML)
- Keycloak with OpenID Connect (OIDC)
- Google™ Identity Platform (OIDC)
- Auth0
- GitHub™ OAuth

Other configurations might work as well, but the list
above summarizes officially supported platforms. In general, 
most of the identity providers that support OIDC or SAML
can integrate with Pachyderm. However, you might have to
perform additional configuration steps.

## Roles

By default, Pachyderm preconfigures one type of users called
`admin`. Users in the `admin` group can perform any
action on the cluster including appointing other admins.
Furthermore, if you do not enable Pachyderm authentication,
you have only one type of users – admin users.

If you have activated access controls, in addition to the initial
`admin` user, Pachyderm associates an Access Control List (ACL)
with each repository. Each ACL can include the following
roles:

- `READER` - users who can consume data from the repo, but cannot edit it.
Readers can execute such commands as `pachctl get file`
`pachctl list file`, as well as create pipelines that use data
from this repo. 
- `WRITER` - users who can read and modify data in the repository by
adding, deleting, or updating the files in the repo. Users with
`WRITER` role can perform such operations as `pachctl put file`,
`pachctl delete commit`, and others.
- `OWNER` - users with `READER` and `WRITER` access who can also
modify the repository's ACL.

## User Account Types

Pachyderm defines the following account types:

* For smaller teams and testing:

  * **GitHub user** is a user account that is associated with
  a GitHub account and logs in through the GitHub OAuth flow. If you do not
  use any third-party identity provider, you use this option. When a user tries
  to log in with a GitHub account, Pachyderm verifies the identity and
  sends a Pachyderm token for that account.

  * **Robot user** is a user account that logs in with a pach-generated authentication
  token. Typically, you create a user in simplified workflow scenarios, such
  as initial SAML configuration.

  * **Pipeline** is an account that Pachyderm creates for
  data pipelines. Pipelines inherit access control from its creator.

  * **SAML user** is a user account that is associated with a SAML identity provider.
  When a user tries to log in through a SAML ID provider, the system
  confirms the identity, associates
  that identity with a SAML identity provider account, and responds with
  the SAML identity provider token for that user. Pachyderm verifies the token,
  drops it, and creates a new internal token that encapsulates the information
  about the user.

  * **OIDC user** is a user that is associated with an OIDC identity provider,
  such as Keycloak, Okta, or other. If you have a user or group configured
  in your identity provider you can give them access to Pachyderm by configuring
  the Pachyderm authentication config.

## Access to Pipelines

In Pachyderm, you do not explicitly grant users access to
pipelines. Instead, pipelines infer access from their input
and output repositories. To update a pipeline, you must have
at least `READER`-level access to all pipeline inputs and at
least `WRITER`-level access to the pipeline output. This is
because pipelines read from their input repos and write
to their output repos, and you cannot grant a pipeline
more access than you have yourself.

An `OWNER`, `WRITER`, or `READER` of a repo can subscribe a
pipeline to that repo. When a user subscribes a pipeline
to a repo, Pachyderm sets that user as an `OWNER` of that
pipeline's output repo. If additional users need access
to the output repository, the initial `OWNER` of a
pipeline's output repo, or an admin, needs to configure
these access rules. To update a pipeline, you must have
`WRITER` access to the pipeline's output repos and `READER`
access to the pipeline's input repos.


## Deactivating Authentication

When an enterprise activation code expires, a
Pachyderm cluster with enable authentication goes into an
`admin-only` state. In this state, only admins have
access to data that is stored in Pachyderm.
This safety measure keeps sensitive data protected, even when
an enterprise subscription becomes stale. As soon as the enterprise
activation code is updated by using the dashboard or CLI, the
Pachyderm cluster returns to its previous state.

When you deactivate access controls on a Pachyderm cluster
by running `pachctl auth deactivate`, the cluster returns
to its original state that including the following changes:

- All ACLs are deleted.
- The cluster returns to being a blank slate in regards to
access control, which means that everyone that can connect
to Pachyderm can access and modify the data in all repos.
- No users are present in Pachyderm, and no one can log in to Pachyderm.

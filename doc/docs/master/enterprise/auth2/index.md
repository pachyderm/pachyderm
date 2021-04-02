# Authentication and Authorization

!!! Note
    Authentication and Authorization are [enterprise features](https://docs.pachyderm.com/latest/enterprise/) that require
    an active enterprise token.

Pachyderm embeds an OpenID Connect identity service based on [**Dex**](https://dexidp.io/docs/) allowing for a vendor-neutral authentication (i.e., a pluggable authentication against many different identity providers). As a result, users can authenticate using their existing credentials from various back-ends, including LDAP, SAML, and other OIDC providers. 


<div class="row">
  <div class="column-2">
    <div class="card-square mdl-card mdl-shadow--2dp">
      <div class="mdl-card__title mdl-card--expand">
        <h4 class="mdl-card__title-text">Learn about Pachyderm's Authentication Flow </h4>
      </div>
      <div class="mdl-card__supporting-text">
        <ul>
            <li><a href="deploy/" class="md-typeset md-link">
            Authentication
            </a>
            </li>
          </ul>
      </div>
  </div>
  <div class="column-2">
    <div class="card-square mdl-card mdl-shadow--2dp">
      <div class="mdl-card__title mdl-card--expand">
        <h4 class="mdl-card__title-text">Users and Access Management</h4>
      </div>
      <div class="mdl-card__supporting-text">
        Test supporting text
      </div>
      <div class="mdl-card__actions mdl-card--border">
          <ul>
            <li><a href="manage/" class="md-typeset md-link">
            Authorization
           </a>
          </li>
       </div>
     </div>
  </div>
</div>

More about Pachyderm Auth flow
Users and Access Management

To configure your IdP with Pachyderm, visit //TODO link to configure Pach in your preferred IdP.


## Activating the Authentication and Authorization features
 


## Deactivating the Authentication and Authorization features

When you deactivate access controls on a Pachyderm cluster
by running `pachctl auth deactivate`, the cluster returns
to its original state that including the following changes:

- All ACLs are deleted.
- The cluster returns to being a blank slate in regards to
access control, which means that everyone that can connect
to Pachyderm can access and modify the data in all repos.
- No users are present in Pachyderm, and no one can log in to Pachyderm.
## Authentication Token Expiration
When an enterprise activation code expires, a
Pachyderm cluster with enable authentication goes into an
`admin-only` state. In this state, only admins have
access to data that is stored in Pachyderm.
This safety measure keeps sensitive data protected, even when
an enterprise subscription becomes stale. As soon as the enterprise
activation code is updated by using the dashboard or CLI, the
Pachyderm cluster returns to its previous state.




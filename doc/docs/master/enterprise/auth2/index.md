# Authentication and Authorization

!!! Note
<<<<<<< HEAD
    User Access Management is an [enterprise feature](https://docs.pachyderm.com/latest/enterprise/) that requires
=======
    Authentication and Authorization are [enterprise features](https://docs.pachyderm.com/latest/enterprise/) that require
>>>>>>> 105ef481c9c08ed53a02ae8ed11aeea066772c0b
    an active enterprise token.

Pachyderm embeds an OpenID Connect identity service based on [**Dex**](https://dexidp.io/docs/) allowing for a vendor-neutral authentication (i.e., a pluggable authentication against many different identity providers). As a result, users can authenticate using their existing credentials from various back-ends, including LDAP, SAML, and other OIDC providers. 

<<<<<<< HEAD
Setting up Pachyderm's User Access Management (also referred to as "Authentication and Authorization" or "Auth" in this documentation) requires to follow the consecutive steps:
1. [Activate the feature](#activate-user-access-management).
1. Connect the IdP of your choice to Pachyderm (Dex). //TODO Link to page
1. Assign IdP users to specific Roles on given Ressources. //TODO Link to page

Any declared IdP users will then be able to log into and access Pachyderm ressources according to the privileges they were granted.


## Activate User Access Management
!!! Note
    Verify the status of your **Enterprise License** before activating the User Access Management feature
    by running `pachctl enterprise get-status`. The command should return an `ACTIVE`
    status along with the expiration date of the Enterprise License.   

To activate Pachyderm's authentication and authorization features,
run the following command in your terminal:

```shell
$ pachctl auth activate 
```
The enablement of the User Access Management **creates
an initial `Root user` and returns a `Root token`**.
This `Root user` (or initial admin) has irrevokable `clusterAdmin` privileges on
Pachyderm's cluster. //TODO Link to list of roles in authorization

**System Response**
```
Pachyderm root token:
54778a770c554d0fb84563033c9cb808
```
!!! Warning 
    You must save the token to a secure location
    to avoid being locked out of your cluster.
    
    When needed, use this token to log back in as this initial admin user:

    ```shell
    pachctl auth use-auth-token
    ```

As a *Root User* (or initial admin), 
you can now configure Pachyderm to work with
the identity management provider (IdP) of your choice.
//TODO Link to Configure an IdP with Pachyderm

## Deactivating User Access Management
Deactivating the User Access Management on a Pachyderm cluster
(as a `clusterAdmin`, run `pachctl auth deactivate`), 
returns the cluster to being a blank slate with regards to
access control.

This specifically implies that everyone that can connect
to Pachyderm is back to being a `clusterAdmin` (can access and modify all data in all repos).


## User Access Management and Enterprise License expiration
When an Enterprise License expires, a
Pachyderm cluster with enabled User Access Management goes into an
`admin-only` state. In this state, only `ClusterAdmins` have
access to the data stored in Pachyderm.

This safety measure keeps sensitive data protected, even when
an enterprise subscription becomes stale. 

As soon as the enterprise
activation code is updated (As a 'clusterAdmin', run `pachctl license activate` and submit your new code), the
=======

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
>>>>>>> 105ef481c9c08ed53a02ae8ed11aeea066772c0b
Pachyderm cluster returns to its previous state.



<<<<<<< HEAD

=======
>>>>>>> 105ef481c9c08ed53a02ae8ed11aeea066772c0b

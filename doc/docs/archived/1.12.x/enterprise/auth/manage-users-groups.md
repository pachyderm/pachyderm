# Manage Users and Groups

If you have users or groups configured in a third-party
identity provider, you can give access to Pachyderm to
those users and groups. The following table summarizes
which providers support user and group authentication:

| Provider        | Users           | Groups          |
| --------------- | --------------- | --------------- |
| GitHub          | &#10004;        | X               |
| Otka (OIDC)     | &#10004;        | &#10004;        |
| Keycloak (OIDC) | &#10004;        | &#10004;        |
| Keycloak (SAML) | &#10004;        | &#10004;        |
| Google (OIDC)   | &#10004;        | X               | 
| Auth0 (OIDC)    | &#10004;        | X               |
| Auth0 (SAML)    | &#10004;        | X               |
## Configure User Access

You can manage user access in the UI and CLI.
For example, you are logged in to Pachyderm as the user `user1`
and have a repository called `test`.  Because the user `user1` created
this repository, `user1` has full `OWNER`-level access to the repo.
You can confirm this in the dashboard by navigating to or clicking on
the repo. Alternatively, you can confirm your access by running the
 `pachctl auth get ...` command:

!!! example

    ```shell
    pachctl auth get dwhitena test
    ```

    **System response:**

    ```shell
    OWNER
    ```

An OWNER of `test` or a cluster admin can then set other user
level of access to the repo by using the `pachctl auth set ...`
command or through the dashboard.

For more information about the roles that you can assign,
see [Roles]().



To manage user access, complete the following steps:

* If you are using the dashboard:

  1. In the dashboard, click **Repo**.
  1. Select the repo to which you want to grant access to your users.
  1. Click **Modify access controls**. 
  1. Add the users to a desired list of `READERs`, `WRITERs`,
  or `OWNERs`.

     For example, to give the GitHub user `user2` `READER`, but not
     `WRITER` or `OWNER`, access to the `test` repository add them
     to the `READER` list.

* If you are using `pachctl`:

  1. Grant a user an access to a repo:

     ```shell
     pachctl auth set <username> (none|reader|writer|owner) <repo>
     ```

     **Example:**

     ```shell
     pachctl auth set user1 reader test
     ```

  1. Verify the ACL for the repo:

     ```shell
     pachctl auth get <repo>
     ```

     **Example:**

     ```shell
     pachctl auth get test
     ```

     **System Response:**

     ```shell
     github:svekars: OWNER
     github:user1: READER
     ```

## Configure Group Access

If you have a group of users configured in an identity provider,
you can grant access to a Pachyderm repository to all users
in that group.

!!! note
    This functionality is experimental and supported only
    through the command line. The changes will not be
    visible in the UI.

To configure group access for SAML providers, you need to set the `group_attibute` in
the `id_providers` field of your authentication config:

**Example:**

   ```shell
   pachctl auth set-config <<EOF
   {
     ...
     "id_providers": [
       {
         ...
         "saml": {
           "group_attribute": "memberOf"
         }
       }
     ],
   }
   EOF
   ```

To configure groups in OIDC providers, you need to add the `groups` scope to
`additional_scopes` in your authentication config:

**Example:**

   ```shell
   pachctl auth set-config <<EOF
   {
     ...
     "id_providers": [
       {
         ...
         "oidc": {
           "additional_scopes": ["groups"]
         }
       }
     ],
   }
   EOF
   ```


!!! note "See also"
    [Configure a SAML User](saml/index.md)

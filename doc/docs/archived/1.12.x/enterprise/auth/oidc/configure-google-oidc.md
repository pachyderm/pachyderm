# Configure Google OpenID Connect

You can use Google® OAuth 2.0 authentication system as an identity
provider for Pachyderm. Google takes care of verifying the identity of
users accessing your Pachyderm cluster. 

## Prerequisites

Before you can configure Pachyderm to work with Keycloak, you need to
have the following components up and running:

* Pachyderm Enterprise 1.11.x or later. The enterprise token must be
  activated by running `echo <your-activation-token> | pachctl enterprise activate`.
  Check the status of your license by running:

  ```shell
  pachctl enterprise get-state
  ```

  For more information, see [Activate Pachyderm Enterprise Edition](../../../deployment/#activate-pachyderm-enterprise-edition).

* A Google account, such as a Gmail account. You need to have access to
  the Google API Console and have a project there. For more information,
  see [Google OpenID Connect documentation](https://developers.google.com/identity/protocols/oauth2/openid-connect#sendauthrequest).

## Configure Google OAuth 2.0

You need to create a project in the Google API Console within an
organization. Some of the IAM features that are discussed in this section
are no available to individual users. This section outlines Pachyderm
specifics for setting up authentication with Google. For more specific
details about the configuration, see [Google OpenID Connect documentation](https://developers.google.com/identity/protocols/oauth2/openid-connect#sendauthrequest).

To set up Google OAuth 2.0, complete the following steps:

1. Go to your project in Google API Console.
1. Click **Credentials**.
1. Click **Configure Consent Screen**.
1. Select a user type as needed for your project.
1. Type the **Application name**.

   This action creates a client within Google OAuth 2.0 authentication system.
   You can fill out other fields as needed, but to authorize with Pachyderm,
   only the application name is required.

1. Save the settings.
1. Go back to **Credentials**.
1. Under **OAuth 2.0 Client IDs**, edit the client that you have created in the
previous step.
1. In the **Authorized redirect URIs** section, add the Pachyderm callback link
in the following format:

   ```shell
   https://<hostname>:30657/authorization-code/callback
   ```

   This is your `redirect_uri`.
   The path cannot include an IP address and must have the OIDC protocol.

1. Click **Save**.
1. Go to [Configure Pachyderm](#configure-pachyderm).

## Configure Pachyderm

After you have completed the steps in [Configure Google OAuth 2.0](#configure-google-oauth-20),
you need to create a Pachyderm authentication config and login as a Google user
to your Pachyderm cluster.

To configure Pachyderm, complete the following steps:

1. Go to the terminal and forward the `pachd` pod to the OIDC port:

   1. Get the `pachd` pod ID:

      ```shell
      kubectl get pod
      ```

      **Example system response:**

      ```shell
      dash-5768cb7d98-j6cgt       2/2     Running   0          4h2m
      etcd-56d897697-xzsqr        1/1     Running   0          4h2m
      keycloak-857c59449b-htg99   1/1     Running   0          4h6m
      pachd-79f7f68c65-9qs8g      1/1     Running   0          4h2m
      ```

   1. Forward the `pachd` pod to the OIDC port:

      **Example:**

      ```shell
      kubectl port-forward pachd-79f7f68c65-9qs8g 30657
      ```

1. Enable Pachyderm authentication:

   ```shell
   pachctl auth activate --initial-admin=robot:admin
   ```

   Pachyderm returns a token.

   **WARNING!** You must save the token to a secure location
   to avoid being locked out of your cluster.

 1. Log in as the admin user with the token you received in the previous
 step:

    ```shell
    pachctl auth use-auth-token
    ```

1. Set up the authentication config:

   ```shell
   pachctl auth set-config <<EOF
   {
             "live_config_version": 2,
             "id_providers": [{
             "name": "google-oauth",
             "description": "oidc-based authentication with Google OAuth 2.0",
             "oidc":{
                     "issuer": "https://accounts.google.com",
                     "client_id": "<client-id>",
                     "client_secret": "<client-secret>",
                     "redirect_uri": "http://<hostname>:30657/authorization-code/callback"
             }
         }]
   }
   EOF
   ```

   You need to replace the following placeholders with relevant values:

   - `issuer` — In ase of Google OAuth 2.0, this will always be
   `https://accounts.google.com`.

   - `client_id` — The Pachyderm **Client ID** in the Google OAuth
   2.0 Credentials page.

   - `client_secret` - The Pachyderm client secret in in the Google OAuth
   2.0 Credentials page.
   - `redirect_uri` - This parameter should match what you have added
   to **Authorized redirect URIs** in the previous section.

1. Log in as the user you have created in the Pachyderm application
or sing in with Google:

   1. Run:

      ```shell
      pachctl auth login
      ```

      You should be prompted to a web-browser. Sign in with your
      Google account.

      You should see the following message printed out in your browser:

      ```
      You are now logged in. Go back to the terminal to use Pachyderm!
      ```

1. In the terminal, check that you are logged in as the Auth0 user:

   ```shell
   pachctl auth whoami
   ```

   **Example of System Response:**

   ```shell
   You are "google-oauth:test@pachyderm.com"
   session expires: 07 Aug 20 16:27 PDT
   ```

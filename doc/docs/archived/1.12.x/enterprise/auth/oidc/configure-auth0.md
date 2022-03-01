# Configure OIDC with OAuth0

[Auth0](https://auth0.com/) is an online authentication platform that
developers can use to log in to various applications.
If Auth0 is your preferred choice of the identity provider,
you can configure Pachyderm with Auth0 so that your
Pachyderm users can log in through Auth0.

# Prerequisites

Before you can configure Pachyderm to work with Keycloak, you need to
have the following components configured:

* Pachyderm Enterprise 1.11.x or later. The enterprise token must be
  activated by running `echo <your-activation-token> | pachctl enterprise activate`.
  Check the status of your license by running:

  ```shell
  pachctl enterprise get-state
  ```

  For more information, see [Activate Pachyderm Enterprise Edition](../../../deployment/#activate-pachyderm-enterprise-edition).

* An account at https://auth0.com . It could be a new or an existing
account, either paid or free.

## Configure Auth0

If you do not have an Auth0 account, you need to sign up for one
at https://auth0.com . Then, you need to add Pachyderm as an application
and configure it to work with Auth0.

To configure Auth0, complete the following steps:

1. Log in to your Auth0 account.
1. In the **Applications**, click **Create Application**.
1. Type the name of your application, such as **Pachyderm**.
1. In the application type, select **Regular Web Application**.
1. Click **Create**.
1. Go to the application settings.
1. Scroll down and click **Show Advanced Settings**.
1. Select **Grant Types**.
1. Verify that **Implicit**, **Authorization Code**, **Refresh Token**, and
**Client Credentials** are selected.

   ![Auth0 Grant Settings](../../../assets/images/s_auth0_grant_settings.png)

1. In the **Allowed Callback URLs**, add the Pachyderm callback link in the
following format:

   ```shell
   http://<ip>:30657/authorization-code/callback
   ```

   The IP address is the address of your Pachyderm host. For example,
   if you are running Pachyderm in Minikube, you can find the IP
   address by running `minikube ip`.

1. Proceed to [Configure Pachyderm Auth](#configure-pachyderm-auth).

## Configure Pachyderm Auth

After you have configured a Pachyderm application in Auth0, you
need to create a Pachyderm OIDC config with the Auth0 parameters.
All the required parameters, such as `client_id`, `client_secret`, 
and othersi, are located on the application settings screen. In addition, OIDC
settings are exposed at https://appication-domain/.well-known/openid-configuration.

To configure Pachyderm Auth, complete the following steps:

1. Check the status of your license by running:

   ```shell
   pachctl enterprise get-state
   ```

   You must have an active enterprise token to proceed.

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

   1. Log in as the admin user with the token you received in the previous
step:

   ```shell
   pachctl auth use-auth-token
   ```

1. Set up the authentication config:

    ```shell
    pachctl auth set-config <<EOF
    {
            "live_config_version": 1,
            "id_providers": [{
            "name": "auth0",
            "description": "oidc-based authentication with Auth0",
            "oidc":{
                    "issuer": "<domain>",
                    "client_id": "<client-id>",
                    "client_secret": "<client-secret>",
                    "redirect_uri": "http://<ip>:30657/authorization-code/callback"
            }
        }]
    }
    EOF
    ```

    You need to replace the following placeholders with relevant values:

    - `issuer` — The domain of your application in Auth0. For example,
    `dev-7vllfmvr.us.auth0.com/`. Note the trailing slash.

    - `client_id` — The Pachyderm **Client ID** in Auth0. The client ID
    consists of alphanumeric characters and can be found on the application
    settings page.

    - `client_secret` - The Pachyderm client secret in Auth0 located
    on the application settings page.
    - `redirect_uri` - This parameter should match what you have added
    to **Allowed Callback URLs** in the previous step.

1. Log in as the user you have created in the Pachyderm application
or sign in with Google:

   1. Run:

      ```shell
      pachctl auth login
      ```

      You should be prompted to a web-browser. Log in as the user you have
      previously created in Auth0 or sign in with Google.

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
   You are "auth0:test@pachyderm.com"
   session expires: 07 Aug 20 14:04 PDT
   ```

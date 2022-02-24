# Configure OpenID Connect with Okta

If [Okta® access management software](https://www.okta.com) 
is your preferred choice of IdP,
you can configure Pachyderm to use Okta as an OpenID Connect (OIDC) 
identity provider using the following steps. 


# Prerequisites

Before you can configure Pachyderm to work with Okta, you need:

* Pachyderm Enterprise 1.11.x or later. The enterprise token must be
  activated by running `echo <your-activation-token> | pachctl enterprise activate`.
  Check the status of your license by running:

      ```shell
      pachctl enterprise get-state
      ```

      For more information, see [Activate Pachyderm Enterprise Edition](../../../deployment/#activate-pachyderm-enterprise-edition).

* An account at https://www.okta.com/login/. 


## Register Pachyderm with Okta

For more detailed step by step instructions, follow this [documentation](https://developer.okta.com/docs/guides/add-an-external-idp/apple/register-app-in-okta/).

1. Sign in to your Okta organization with your administrator account.
1. From the Admin Console side navigation, click **Applications > Applications**.
1. Click **Add Application**.
1. Click **Create New App** (or search for your existing app).
1. Select **Platform: Web** and sign-on method **OpenID Connect**.
1. Click **Create**.
1. Type the name of your application, such as **Pachyderm**.
1. Add the following Login redirect URI. 
      ```shell
      http://<ip>:657/authorization-code/callback
      ```
      Note: Your port number should be whatever is routing to pachd:657.

      The IP address is the address of your Pachyderm host. For example,
      if you are running Pachyderm in Minikube, you can find the IP
      address by running `minikube ip`.

1. Click **Save**
1. Click **Edit** to change the General Settings pane. In the Allowed grant types section, enable **Implicit**, **Authorization Code**, **Refresh Token**, and **Client Credentials**.
1. Click **Save**
1. On the Assignments tab, click **Assign** to assign the app integration to any user or group in your org. Click **Done** when the assignments are complete.


## Configure Pachyderm Auth

After you have configured a Pachyderm application in Okta, you
need to create a Pachyderm OIDC config with the Okta parameters.
All the required parameters, such as `client_id`, `client_secret`, 
and others, are located on the App General tab.

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
      "name": "okta",
      "description": "oidc-based authentication with Okta",
      "oidc":{
      "issuer": "https://",
      "client_id": "",
      "client_secret": "",
      "redirect_uri": "your redirect URI",
      ignore_email_verified: true
      }
      }]
      }
      EOF
    ```

    You need to replace the following placeholders with relevant values:

    - `issuer` — The domain of your application in Okta. For example,
    `{yourOktaDomain}/`. Note the trailing slash.

    - `client_id` — The Pachyderm **Client ID** in Okta. 

    - `client_secret` - The Pachyderm client secret in Okta. 

    - `redirect_uri` - This parameter should match what you have added
    to **redirect URI** in the previous step.

1. Log in as the user you have created in the Pachyderm application
or sign in with Google:

   1. Run:

      ```shell
      pachctl auth login
      ```

      You should be prompted to a web-browser. Log in as the user you have
      previously created in Okta or sign in with Google.

    You should see the following message printed out in your browser:

    ```
    You are now logged in. Go back to the terminal to use Pachyderm!
    ```

1. In the terminal, check that you are logged in as the Okta user:

      ```shell
      pachctl auth whoami
      ```

      **Example of System Response:**

      ```shell
      You are "okta:test@pachyderm.com"
      session expires: 07 Aug 20 14:04 PDT
      ```

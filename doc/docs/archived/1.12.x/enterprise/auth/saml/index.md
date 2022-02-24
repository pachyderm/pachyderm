# Configure SAML 


This section walks you through the configuration an SAML IdP for Pachyderm.
We used Auth0 for this example.

!!! Note
    SAML is not supported with Okta. However, you can
    configure Okta with [OIDC auth config](../oidc/configure-with-okta).


When starting out, we recommend running Pachyderm locally, as
mistakes in this configuration could lock you out of your cluster for good.
Experimenting with authentication options is easier and
safer locally.

## Before you start

Activate Pachyderm enterprise and Pachyderm auth:

```shell
  echo <your-activation-token> | pachctl enterprise activate
  pachctl auth activate --initial-admin=robot:admin
```

You now have admin privileges in the cluster (run `pachctl auth whoami`, as shown below, to
confirm)

Pachyderm is ready to authenticate & authorize users.

What the `--initial-admin` flag does:

1. Pachyderm requires at least one cluster admin when auth is activated.
1. Pachyderm's authentication is built around GitHub by default. Without this flag, Pachyderm asks the caller to go through an OAuth flow with GitHub, and then at the conclusion, makes the caller the cluster admin. Then whoever activated Pachyderm's auth system can modify it by re-authenticating via GitHub and performing any necessary actions.
1. To avoid the OAuth flow, it is also possible to make the initial cluster admin a "robot user". This is what `--initial-admin=robot:<something>` does. Pachyderm will print out a Pachyderm token that authenticates the holder as this robot user. At any point, you can authenticate as this robot user by running the following command:

    ```shell
    pachctl auth use-auth-token
    ```
    **System response:**

    ```shell
    Please paste your Pachyderm auth token:
    <paste robot token emitted by "pachctl auth activate --initial-admin=robot:admin">
    ```

    ```shell
    pachctl auth whoami
    ```

    **System response:**
   ```shell
    You are "robot:admin"
    You are an administrator of this Pachyderm cluster
    ```

The rest of this example assumes that your Pachyderm cluster is running in
minikube, and you are accessing it via `pachctl`'s port forwarding. Many of the
SAML service provider URLs below are set to some variation of `localhost`,
which will only work if you're using port forwarding and your browser is able
to access Pachyderm via `localhost` on the port forwarder's usual ports.

## Create an IdP test app

Configure Auth0 as SAML IdP for Pachyderm:

For more detailed step by step instructions,
follow this [documentation](https://auth0.com/learn/saml-identity-provider/).

1. First sign up for a free account.
1. Navigate to **Applications** on the dashboard, choose your type of application,
click on the **Addons** tab.
1. Follow those [instructions](https://auth0.com/docs/protocols/saml-protocol/configure-auth0-as-saml-identity-provider#configure-auth0-as-idp)
1. set your **Application Callback URL** (enter the URL of Pachyderm to which the SAML assertions should be sent after Auth0 has authenticated the user)
   ```shell
   http://<ip>:657/authorization-code/callback
   ```
   Note: Your port number should be whatever is routing to pachd:657.
   The IP address is the address of your Pachyderm host. For example,
   if you are running Pachyderm in Minikube, you can find the IP
   address by running `minikube ip`.
1. Click **Enable**

Once created, download **the IdP Metadata URL** associated with the test app.

## Write Pachyderm config

Setting an **auth config** is what enables SAML in Pachyderm
(specifically, it enables Pachyderm's Assertion Consumer Service). 
Below is an example config that will
allow users to authenticate in your Pachyderm cluster using the app above.

```shell
# Lookup current config version--pachyderm config has a barrier to prevent
# read-modify-write conflicts between admins
live_config_version="$(pachctl auth get-config | jq .live_config_version)"
live_config_version="${live_config_version:-0}"
```

Set the Pachyderm config:

```shell
pachctl auth set-config <<EOF
{
  # prevent read-modify-write conflicts by explicitly specifying live version
  "live_config_version": ${live_config_version},

  "id_providers": [
    {
      "name": "My test app",
      "description": "IdP test app",
      "saml": {
        "metadata_url": <app metadata URL>,
        "group_attribute": "memberOf" # optional: enable group support
      }
    }
  ],

  "saml_svc_options": {
    # These URLs work if using pachctl port-forward
    "acs_url": "http://localhost:30654/saml/acs",
    "metadata_url": "http://localhost:30654/saml/metadata",
    "dash_url": "http://localhost:30080/auth/autologin",
    "session_duration": "8h",
  }
}
EOF

```

## Logging In
Currently Pachyderm only supports IdP-initiated authentication. To proceed,
configure your app to point to the Pachyderm ACS
(`http://localhost:30654/saml/acs` if using `pachctl`'s port forwarding), then
sign in via the new Auth0 app in your dashboard.

After clicking on the test app, your browser will do a SAML authentication
handshake with your pachyderm cluster, and you will arrive at your Pachyderm
dashboard fully authenticated. To log in with the Pachyderm CLI, get a One-Time
Password from the Pachyderm dash, and then run `pachctl auth login
--code=<one-time password>` in your terminal.

!!! Note
    It may be useful to enable a collection of *debug logs*. To do so,
    add the option `"debug_logging": true` to `"saml_svc_options"`:

    ```json
      pachctl auth set-config <<EOF
      {
        ...
        "saml_svc_options": {
          ...
          "debug_logging": true
        }
      }
      EOF
    ```

## Groups
If your SAML ID provider supports setting group attributes, 
you can use groups to manage access in Pachyderm with the `"group_attribute"` 
in the IDProvider field of the auth config:

```json
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

Then, try:
```shell
  pachctl create repo group-test
  pachctl put file group-test@master -f some-data.txt
  pachctl auth set group/saml:"Test Group" reader group-test
```

Elsewhere:
```shell
  pachctl auth login --code=<auth code>
  pachctl get file group-test@master:some-data.txt # should work for members of "Test Group"
```


## Configuring SAML

Pachyderm supports Security Assertion Markup Language
(SAML) providers to authenticate to Pachyderm. This allows
you to configure Single Sign-On (SSO) by using
browser tokens.

This configuration was performed by using minikube with
port-forwarding enabled and
might not be compatible with other environments.
Many of the SAML service provider URLs below are set to some
variation of `localhost`, which only work if you are using port
forwarding and your browser is able
to access Pachyderm via `localhost` on the port forwarder's usual ports.

This guide walks you through an example of using SAML in Pachyderm,
including the following topics:

1. Activate Pachyderm enterprise and Pachyderm auth.
2. Configure Pachyderm's auth system to enable its SAML ACS, receive SAML
   assertions, and allow us to log in by using Okta.
3. Log in to both the dashboard and CLI.

## Activate Pachyderm Enterprise

When you just starting to use Pachyderm, we **highly**
recommend that you run Pachyderm in Minikube, because
mistakes in this configuration might lock you out of
your cluster.

Because Pachyderm Enterprise requires authentication, you
need to either enable authentication as described in (#auth)
or use the `--initial-admin` flag for a simplified workflow.

By default, Pachyderm's authentication is built around GitHub.
Without the `--initial-admin` flag, Pachyderm asks the caller
to go through an OAuth flow with GitHub, and then at the
conclusion, makes the caller the cluster admin. Then, whoever
activated Pachyderm's authentication system can modify it by
re-authenticating by using GitHub and by performing any
necessary actions

The `--initial-admin=robot:<something>` flag enables you to
avoid the OAuth flow by making the initial cluster admin a
"robot user". When you run the authentication command with
the `--initial-admin` flag, Pachyderm prints out a
Pachyderm token that authenticates the holder as
the robot user. At any point, you can authenticate as this
robot user by running the following command.

```bash
$ pachctl auth use-auth-token
Please paste your Pachyderm auth token:
<paste robot token emitted by "pachctl auth activate --initial-admin=robot:admin">
$ # you are now robot:admin, cluster administrator
 ```

To activate Pachyderm Enterprise and Pachyderm OAuth,
complete the following steps:

1. Activate the Pachyderm Enterprise:

   ```bash
   pachctl enterprise activate <enterprise code>
   ```

1. Activate OAuth by using the `--initial-admin` flag:

   ```bash
   pachctl auth activate --initial-admin=robot:admin
   ```

## Create an IdP test app

This example uses an ID provider (IdP) called Okta.
Here is an example configuration for an Okta test
app that authenticates Okta users with Pachyderm:

![Okta test app config](https://raw.githubusercontent.com/pachyderm/pachyderm/handle_requests_crewjam/doc/auth/okta_form.png)

Once created, you can get the IdP Metadata URL associated with the test Okta
app here:

![Metadata image](https://raw.githubusercontent.com/pachyderm/pachyderm/handle_requests_crewjam/doc/auth/IdPMetadata_highlight.png)

## Configuring the Pachyderm Authentication

Setting an auth config is what enables Pachyderm's SAML Assertion Consumer
Service (ACS), an HTTP endpoint to which users forward SAML assertions.
Below is an example configuration that enables users
to authenticate in their Pachyderm cluster by using the
Okta app.

To configure the Pachyderm configuration file, complete the following steps:

1. Lookup current config version--pachyderm config has a barrier to prevent
read-modify-write conflicts between admins:

```bash
live_config_version="$(pachctl auth get-config | jq .live_config_version)"
live_config_version="${live_config_version:-0}"
```

1. Set the Pachyderm config:

```bash
pachctl auth set-config <<EOF
{
  # prevent read-modify-write conflicts by explicitly specifying live version
  "live_config_version": ${live_config_version},

  "id_providers": [
    {
      "name": "okta",
      "description": "Okta test app",
      "saml": {
        "metadata_url": <okta app metadata URL>,
        "group_attribute": "memberOf" # optional: enable group support
      }
    }
  ],

  "saml_svc_options": {
    # These URLs work if using pachctl port-forward
    "acs_url": "http://localhost:30654/saml/acs",
    "metadata_url": "http://localhost:30654/saml/metadata",
    "dash_url": "http://localhost:30080/auth/autologin",
  }
}
EOF
```

## Logging In

Currently Pachyderm only supports IdP-initiated authentication. To proceed,
configure your Okta app to point to the Pachyderm ACS
(`http://localhost:30654/saml/acs` if using `pachctl`'s port forwarding), then
sign in via the new Okta app in your Okta dashboard.

After clicking on the test Okta app, your browser will do a SAML authentication
handshake with your pachyderm cluster, and you will arrive at your Pachyderm
dashboard fully authenticated. To log in with the Pachyderm CLI, get a One-Time
Password from the Pachyderm dash, and then run `pachctl auth login
--code=<one-time password>` in your terminal.

### Groups
If your SAML ID provider supports setting group attributes, you can use groups to manage access in Pachyderm with the `"group_attribute"` in the IDProvider field of the auth config:
```
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
```
pachctl create repo group-test
pachctl put file group-test@master -f some-data.txt
pachctl auth set group/saml:"Test Group" reader group-test
```
Elsewhere:
```
pachctl auth login --code=<auth code>
pachctl get file group-test@master:some-data.txt # should work for members of "Test Group"
```


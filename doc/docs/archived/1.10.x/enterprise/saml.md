# Overview

This guide walks you through an example of using Pachyderm's SAML support, including the
following:

1. Activate Pachyderm enterprise and Pachyderm auth.
2. Configure Pachyderm's auth system to enable its SAML ACS, receive SAML
   assertions, and allow you to log in by using the Okta® access management
   software.
3. Log in to both the dash and CLI.

## Activation

When starting out, we **highly** recommend running Pachyderm in Minikube, as
mistakes in this configuration could lock you out of your cluster.

To activate Pachyderm enterprise and Pachyderm auth:

```
echo <your-activation-token> | pachctl enterprise activate
pachctl auth activate --initial-admin=robot:admin
```

At this point, Pachyderm is ready to authenticate & authorize users.

What the `--initial-admin` flag does is this:
1. Pachyderm requires there to be at least one cluster admin if auth is
   activated
2. Pachyderm's authentication is built around GitHub by default. Without this
   flag, Pachyderm asks the caller to go through an OAuth flow with GitHub, and
   then at the conclusion, makes the caller the cluster admin. Then whoever
   activated Pachyderm's auth system can modify it by re-authenticating via
   GitHub and performing any necessary actions
3. To avoid the OAuth flow, though, it's also possible to make the initial
   cluster admin a "robot user". This is what
   `--initial-admin=robot:<something>` does.
4. Pachyderm will print out a Pachyderm token that authenticates the holder as
   this robot user. At any point, you can authenticate as this robot user by
   running the following command:

   ```shell
   pachctl auth use-auth-token
   ```

   **System response:**

   ```shell
   Please paste your Pachyderm auth token:
   <paste robot token emitted by "pachctl auth activate --initial-admin=robot:admin">
   # you are now robot:admin, cluster administrator
   ```

The rest of this example assumes that your Pachyderm cluster is running in
minikube, and you're accessing it via `pachctl`'s port forwarding. Many of the
SAML service provider URLs below are set to some variation of `localhost`,
which will only work if you're using port forwarding and your browser is able
to access Pachyderm via `localhost` on the port forwarder's usual ports.

## Create IdP test app
The ID provider (IdP) that this example uses is Okta. Here is an example
configuration for an Okta test app that authenticates Okta users
with Pachyderm:

![Okta test app config](https://raw.githubusercontent.com/pachyderm/pachyderm/handle_requests_crewjam/doc/auth/okta_form.png)

Once created, you can get the IdP Metadata URL associated with the test Okta
app here:

![Metadata image](https://raw.githubusercontent.com/pachyderm/pachyderm/handle_requests_crewjam/doc/auth/IdPMetadata_highlight.png)

## Write Pachyderm config
Broadly, setting an auth config is what enables SAML in Pachyderm
(specifically, it enables Pachyderm's ACS). Below is an example config that will
allow users to authenticate in your Pachyderm cluster using the Okta app above.
Note that this example assumes

```
# Lookup current config version--pachyderm config has a barrier to prevent
# read-modify-write conflicts between admins
live_config_version="$(pachctl auth get-config | jq .live_config_version)"
live_config_version="${live_config_version:-0}"
# Set the Pachyderm config
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


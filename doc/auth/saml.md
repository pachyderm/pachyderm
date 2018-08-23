## Overview

This guide will walk through testing Pachyderm's experimental SAML support.
These features aren't integrated into mainline Pachyderm yet and aren't
available in any official releases. This will describe the process of:

1. Activating Pachyderm enterprise and Pachyderm auth
1. Configuring Pachyderm's auth system to enable its SAML ACS and receive SAML
    assertions
1. Logging in to both the dash and CLI
1. Enabling debug logging in case anything goes wrong

## Activation

For testing, we **highly** recomment running Pachyderm in Minikube, in case any
early bugs make it necessary to restart the cluster.

To activate Pachyderm enterprise and Pachyderm auth:

```
pachctl enterprise activate <enterprise code>
pachctl auth activate --initial-admin=robot:admin
```

At this point, Pachyderm is ready to authenticate & authorize users.

What the `--initial-admin` flag does is:
1. Pachyderm requires there to be at least one cluster admin if auth is
   activated
1. Pachyderm's authentication is build around GitHub by default. Without this
   flag, Pachyderm asks the caller to go through an OAuth flow with GitHub, and
   then at the conclusion, makes the caller the cluster admin. Then whoever
   activated Pachyderm's auth system can modify it by re-authenticating via
   GitHub and performing any necessary actions
1. To avoid the OAuth flow, though, it's also possible to make the initial
   cluster admin a "robot user". This is what
   `--initial-admin=robot:<something>` does.
1. Pachyderm will print out a Pachyderm token that authenticates the holder as
   this robot user. At any point, you can authenticate as this robot user by
   running
   ```
   $ pachctl auth use-auth-token
   Please paste your Pachyderm auth token:
   <paste robot token emitted by "pachctl auth activate --initial-admin=robot:admin">
   $ # you are now robot:admin, cluster administrator
   ```

## Create IdP test app
Here is the configuration I used for an Okta test app that authenticates Okta users
with Pachyderm:
![Okta test app config](https://github.com/pachyderm/pachyderm/blob/handle_requests_crewjam/doc/auth/okta_form.png)

Once created, I was able to get the IdP Metadata URL associated with this app here:
![Metadata image](https://github.com/pachyderm/pachyderm/blob/handle_requests_crewjam/doc/auth/IdPMetadata_highlight.png)

## Write Pachyderm config
This is what enables Pachyderm ACS. See inline comments

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
        "group_attribute": "memberOf" # optional: enable experimental group support
      }
    }
  ],

  "saml_svc_options": {
    # These URLs work if using pachctl port-forward
    "acs_url": "http://localhost:30654/saml/acs",
    "metadata_url": "http://localhost:30654/saml/metadata",
    "dash_url": "http://localhost:30080/auth/autologin",

    # optional: enable verbose logging from auth, to fix bugs
    "debug_logging": true
  }
}
EOF
```

## Logging In
Currently Pachyderm only supports IdP-initiated authentication. Configure
an Okta app to point to the Pachyderm ACS
(`http://localhost:30654/saml/acs` if using `pachctl port-forward`), then
sign in via the new Okta app

This should allow you to log in at the Pachyderm dash. To log in with the
Pachyderm CLI, get a One-Time Password from the Pachyderm dash, and then
run `pachctl auth login --code=<one-time password>` in your terminal.

## Other features
### Debug Logging
If we run into issues while deploying this, it may be useful to enable
a collection of debug logs that we added during development. To do so,
add the option `"debug_logging": true` to `"saml_svc_options"`:
```
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

### Groups
Pachyderm has very preliminary, experimental support for groups. While they won't
appear in ACLs in the dash (and may have other issues), you can experiment using
the CLI by setting `"group_attribute"` in the IDProvider field of the auth config:
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
pachctl create-repo group-test
pachctl put-file group-test master -f some-data.txt
pachctl auth set group/saml:"Test Group" reader group-test
```
Elsewhere:
```
pachctl auth login --code=<auth code>
pachctl get-file group-test master /some-data.txt # should work for members of "Test Group"
```


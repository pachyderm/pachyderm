## Overview

This guide will walk through an example of using Pachyderm's experimental SAML
support. We'll describe:

1. Authenticating via a SAML IdP in Pachyderm
1. Authenticating in the CLI
1. Authorizing a user or group to access data

## Setup

Follow the instructions in [saml_setup](saml_setup.md) to enable auth in a
Pachyderm cluster and connect it to a SAML ID provider. Then, we'll authenticate
as a cluster admin in one console and set up our [open CV
demo](../examples/opencv/README.md).

## Create IdP test app
Here is the configuration I used for an Okta test app that authenticates Okta users
with Pachyderm:
![Okta test app config](images/okta_form.png)

Once created, I was able to get the IdP Metadata URL associated with this app here:
![Metadata image](images/IdPMetadata_highlight.png)

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


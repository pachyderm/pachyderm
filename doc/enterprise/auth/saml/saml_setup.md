# Configure a SAML application

Pachyderm supports Security Assertion Markup Language
(SAML) providers to authenticate with Pachyderm which enables
you to configure Single Sign-On (SSO) by using browser tokens.

To complete the steps in this section, use the Minikube
configuration with enabled port-forwarding. Many of the SAML
service provider URLs below are set to a variation of localhost
that only works if you are using port forwarding and your browser
can access Pachyderm through localhost on the port forwarder’s usual ports.

Because the examples in this section describe the configuration that
uses Okta as an Identity Provider, you need to have Okta installed
and configured before you start executing the steps in this section.

This section walks you through an example of using SAML in Pachyderm,
including the following topics:

1. Activate Pachyderm enterprise and Pachyderm auth.
1. Configure the Pachyderm’s auth system to enable its
SAML Assertion Consumer Service (ACS) and SSO by using Okta,
as well as receive SAML assertions.
1. Log in to both the dashboard and the CLI.

## Activate Pachyderm Enterprise

When you are just starting to use Pachyderm, we highly recommend that
you run Pachyderm in Minikube because mistakes in this configuration
might lock you out of your cluster.

Because Pachyderm Enterprise requires authentication, you need to
either enable authentication as described in (#auth) or use the
`--initial-admin` flag for a simplified workflow.

By default, Pachyderm’s authentication is built around GitHub.
Without the `--initial-admin` flag, Pachyderm asks the caller to go
through an OAuth flow with GitHub, and then at the conclusion,
makes the caller the cluster admin. Then, the user that activated the
Pachyderm’s authentication system can modify it by re-authenticating
with a GitHub account.

The `--initial-admin=robot:<something>` flag enables you to avoid
the OAuth flow by making the initial cluster admin a *robot user*.
When you run the authentication command with the `--initial-admin`
flag, Pachyderm prints out a Pachyderm token that authenticates
the holder as the robot user. At any point, you can authenticate
as this robot user by running the following command:

```bash
$ pachctl auth use-auth-token
Please paste your Pachyderm auth token:
<paste robot token emitted by "pachctl auth activate --initial-admin=robot:admin">
$ # you are now robot:admin, cluster administrator
```

To activate Pachyderm Enterprise and Pachyderm OAuth, complete the following
steps:

1. Activate Pachyderm Enterprise:

   ```bash
   pachctl enterprise activate <enterprise code>
   ```

1. Activate OAuth by using the `--initial-admin` flag:

   ```bash
   pachctl auth activate --initial-admin=robot:admin
   ```

## Create an IdP test application

This example describes a configuration with an external Identity Provider
called Okta.

To create an IdP test application, complete the following steps:

1. In the Okta UI, configure a test Okta app.

   This screenshot demonstrates an example configuration for an Okta
   test app that authenticates Okta users with Pachyderm:

   ![Okta test app config](https://raw.githubusercontent.com/pachyderm/pachyderm/handle_requests_crewjam/doc/auth/okta_form.png)

1. Associate the IdP Metadata URL with the test Okta app as shown
   on the screenshot below::

![Metadata image](https://raw.githubusercontent.com/pachyderm/pachyderm/handle_requests_crewjam/doc/auth/IdPMetadata_highlight.png)

## Set up the Pachyderm Authentication config file

Setting an auth config is what enables Pachyderm's SAML ACS, an
HTTP endpoint to which users forward SAML assertions.
Below is an example configuration that enables users
to authenticate in their Pachyderm cluster by using the
Okta app.

To configure the Pachyderm configuration file, complete the following steps:

1. Look up current configuration version. Pachyderm config has a barrier
that prevents read-modify-write conflicts between admins:

   ```bash
   live_config_version="$(pachctl auth get-config | jq .live_config_version)"
   live_config_version="${live_config_version:-0}"
   ```

1. Set the Pachyderm authentication config:

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

## Log in to Dashboard

Currently, Pachyderm only supports IdP-initiated authentication.

To log in, follow these steps:

1. Configure your Okta app to point to the Pachyderm ACS. If you use
port-forwarding, user `http://localhost:30654/saml/acs`.

1. In the Okta dashboard, sign in by using the new Okta app.
   When you click on the test Okta app, your browser performs a
   SAML authentication handshake with your Pachyderm cluster fully
   authenticates Pachyderm with Okta.

To log in with the Pachyderm CLI, complete the following steps:

1. In the Pachyderm dashboard, get a One-Time password.

1. In terminal, run:

   ```bash
   pachctl auth login --code=<one-time-password>
   ```

## Configure ID Provider Groups

If your SAML ID provider supports setting group attributes, you can
use groups to manage access in Pachyderm with the `"group_attribute"`
parameter in the `id_providers` field of the authentication config
file.

To configure ID provider groups, complete the following steps:

1. Specify the ID provider in the auth config file:

   ```bash
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

1. Create a repository:

   ```bash
   pachctl create repo group-test
   ```

1. Upload a file to the repository:

   ```bash
   pachctl put file group-test@master -f some-data.txt
   ```

1. Set the access privileges that a group has to the
the specified repository:

   ```bash
   pachctl auth set <groupname> <none|reader|writer|owner> <repository>
   ```

   The following examples give reader permissions to the `saml` group
   in the `group-test` repository:

   ```bash
   pachctl auth set group/saml:"Test Group" reader group-test
   ```

To verify your configuration, complete the following steps:

1. Log in to `pachctl` with the one-time password from the dashboard:

   ```bash
   pachctl auth login --code=<one-time-pass>
   ```
1. Get the list of files from the `group-test` repository:

   ```bash
   pachctl get file group-test@master:some-data.txt
   ```
   This command should work for members of the `Test Group`.

## Troubleshoot the auth config

To debug the authentication configuration, add `"debug_logging": true`
the config file:

```bash
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

## Use SAML

This section walks you through an example of using Pachyderm's experimental SAML
support. We'll describe:

1. Authenticating via a SAML ID Provider
1. Authenticating in the CLI
1. Authorizing a user or group to access data

## Setup

Follow the instructions in [Configure SAML](saml_setup.md) to enable auth in a
Pachyderm cluster and connect it to a SAML ID provider. Then, we'll authenticate
as a cluster admin in one console and set up our [open CV
demo](https://github.com/pachyderm/pachyderm/blob/1.13.x/examples/opencv/README.md).

In the CLI, that would look like:

```
(admin)$ pachctl auth use-auth-token
```

**System response:**

```shell
Please paste your Pachyderm auth token:
<auth token>
```

```shell
(admin)$ pachctl auth whoami
```

**System response:**

```shell
You are "robot:admin"
You are an administrator of this Pachyderm cluster
```

Create a repository, pipelines, and put files into the
repository.

```shell
(admin)$ pachctl create repo images
(admin)$ pachctl create pipeline -f examples/opencv/edges.json
(admin)$ pachctl create pipeline -f examples/opencv/montage.json
(admin)$ pachctl put file images@master -i examples/opencv/images.txt
(admin)$ pachctl put file images@master -i examples/opencv/images2.txt
```

View the list of existing repositories:

```shell
(admin)$ pachctl list repo
```

**System response:**

```shell
NAME    CREATED       SIZE (MASTER) ACCESS LEVEL
montage 2 minutes ago 1.653MiB      OWNER
edges   2 minutes ago 133.6KiB      OWNER
images  2 minutes ago 238.3KiB      OWNER
```

View the list of jobs:

```shell
(admin)$ pachctl list job
```

**System response:**

```shell
ID                               OUTPUT COMMIT                            STARTED       DURATION  RESTART PROGRESS  DL       UL       STATE
023a478b16e849b4996c19632fee6782 montage/e3dd7e9cacc5450c92e0e62ab844bd26 2 minutes ago 8 seconds 0       1 + 0 / 1 371.9KiB 1.283MiB success
fe8b409e0db54f96bbb757d4d0679186 edges/9cc634a63f794a14a78e931bea47fa73   2 minutes ago 5 seconds 0       2 + 1 / 3 181.1KiB 111.4KiB success
152cb8a0b0854d44affb4bf4bd57228f montage/82a49260595246fe8f6a7d381e092650 2 minutes ago 5 seconds 0       1 + 0 / 1 79.49KiB 378.6KiB success
86e6eb4ae1e74745b993c2e47eba05e9 edges/ee7ebdddd31d46d1af10cee25f17870b   2 minutes ago 4 seconds 0       1 + 0 / 1 57.27KiB 22.22KiB success
```

## Authenticating via a SAML ID Provider (in the dashboard)
Before authenticating, navigating to the dash will yield a blank screen:

![Blocked-out dash](../assets/images/saml_log_in.png)

Even through the dash suggests logging in via GitHub, we will log in using a
SAML IdP (which has hopefully already been configured). To see your Pachyderm
DAG, navigate to your SAML ID provider and sign in to your Pachyderm cluster
there (currently Pachyderm only supports IdP-initiate SAML authentication).

![SSO image](../assets/images/saml_okta_with_app.png)

Once you've authenticated, you'll be redirected to the Pachyderm dash (the
redirect URL is configured in the Pachyderm auth system). You'll be given the
opportunity to generate a one-time password (OTP), though you can always do this
later from the settings panel.

![Dash logged in](../assets/images/saml_successfully_logged_in.png)

After closing the OTP panel, you'll be able to see the Pachyderm DAG, but you
may not have access to any of the repos inside (a repo that you cannot read is
indicated by a lock symbol):

![Dash with locked repos](../assets/images/saml_dag.png)

## Authenticating in the CLI
After authenticating in the dash, you'll be given the opportunity to generate a
one-time password (OTP) and sign in on the CLI. You can also generate an OTP
from the settings panel:

![OTP Image](../assets/images/saml_display_otp.png)

```
(user)$ pachctl auth login --code auth_code:73db4686e3e142508fa74aae920cc58b
(user)$ pachctl auth whoami
```

**System response:**

```shell
You are "saml:msteffen@pachyderm.io"
session expires: 14 Sep 18 20:55 PDT
```

Note that this session expires after 8 hours. The duration of sessions is
configurable in the Pachyderm auth config, but it's important that they be
relatively short, as SAML group memberships are only updated when users sign in.
If a user is removed from a group, they'll still be able to access the group's
resources until their session expires.

## Authorizing a user or group to access data

First, we'll give the example of an admin granting a user access. This can be
accomplished on the CLI like so:

```
(admin)$ pachctl auth set saml:msteffen@pachyderm.io reader images
```

Now, the `images` repo is no longer locked when that user views the DAG:

![Unlocked images repo image](../assets/images/saml_dag_images_readable.png)

At this point, you can click on the `images` repo and preview data inside:

![Unlocked images repo image](../assets/images/saml_dag_reading_from_images.png)

Likewise, you can grant access to repos via groups. You'll need a SAML ID
provider that supports group attributes, and you'll need to put the name of that
attribute in the Pachyderm auth config. Here, we'll grant access to the Everyone
group:

```
(admin)$ pachctl auth set group/saml:Everyone owner edges
```

Now, the edges repo is also not locked:

![Unlocked edges repo](../assets/images/saml_dag_images_and_edges_readable.png)

Also, becase `msteffen@pachyderm.io` has OWNER provileges in the `edges` repo
(via the Everyone group), the ACL for `edges` can be edited.
`msteffen@pachyderm.io` will use OWNER privileges gained via the Everyone group
to add `msteffen@pachyderm.io` (the user principal) directly to that ACL:

![Adding user to ACL image](../assets/images/saml_editing_acl.png)

this change is reflected in the CLI as well:

```shell
(admin)$ pachctl auth get edges
```

**System response:**

```shell
pipeline:edges: WRITER
pipeline:montage: READER
group/saml:Everyone: OWNER
saml:msteffen@pachyderm.io: READER
robot:admin: OWNER
```

## Conclusion

This is just an example of Pachyderm's auth system, meant to illustrate the
general nature of available features. Hopefully, it clarifies whether Pachyderm
can meet your requirements.

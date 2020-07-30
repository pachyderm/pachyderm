# Automated Deferred Processing

[Deferred processing](https://docs.pachyderm.com/latest/how-tos/deferred_processing/) 
is a Pachyderm technique for controlling when data gets processed.
Deferred processing uses branches to prevent pipelines from triggering on every input commit.
This example shows how to automate the movement of those branches,
by using a cron pipeline.
The Makefile in this example,
along with the explanations provided in this document,
should give you a good start on implementing this in your Pachyderm cluster
with or without access controls activated.

In this example, we will cover:

1. Creating a cron pipeline that will move the master branch to the commit in another repo periodically
1. Adding an authentication token to allow it to work when access controls are activated


## Prerequisites

Before you start working on this example, 
you should understand deferred processing by reading the documentation
and trying the [deferred processing example](../deferred_processing_plus_transactions).
That example is used extensively here.

To create and update branch labels, 
the `branch-mover` pipeline uses `pachctl` to send commands to Pachyderm's `pachd`.
The `branch-mover` pipeline, since it's embedded in Pachyderm itself, 
will need a configuration to talk to `pachd` 
and, if access controls are activated, credentials to authenticate itself.
 
For a Pachyderm cluster with activated access controls,
this example demonstrates how to create a Pachyderm authentication token,
load the token into a Kubernetes secret provisioned through `pachctl`, 
and use `transform.secrets` in the pipeline spec,
which both mounts the secret as a Kubernetes volume
and creates an environment variable for use by the pipeline.
If you are unfamiliar with those things,
you might want to refer to the following documentation as you work through the example.

* [Pachyderm access controls and authentication documentation](https://docs.pachyderm.com/latest/enterprise/auth/)
* [Kubernetes documentation on Secrets](https://kubernetes.io/docs/concepts/configuration/secret/)
* The [pachctl create secret](https://docs.pachyderm.com/latest/reference/pachctl/pachctl_create_secret/) command
* [transform.secret in the pipeline specification](https://docs.pachyderm.com/latest/reference/pipeline_spec/)

Before you can start working on this example, make sure you have the following prerequisites:

* You need to have Pachyderm v1.9.8 or later installed on your computer or cloud platform. 
  See [Deploy Pachyderm](https://docs.pachyderm.com/latest/deploy-manage/deploy/).
* Basic familiarity with Makefiles and  Unix shell scripting
* The [jq utility](https://stedolan.github.io/jq/manual/) for transforming json files in shell scripts

## Pipelines

This example uses the same DAG as in the deferred processing example,
with the addition of a cron pipeline 
for periodically moving the `dev` branch to `master`.

For details on the deferred processing example DAG,
see [the Deferred Processing example](../deferred_processing_plus_transactions).

### Branch mover without access controls

If you do not have access controls enabled in your Pachyderm cluster, 
use the instructions in this section. 
Otherwise, proceed to [Branch mover with access controls](#branch_mover_with_access_controls).

The cron pipeline is called `branch-mover`. 
By default,
it is configured to run every minute, 
per its tick input:

```
  "input": {
    "cron": {
        "name": "tick",
        "spec": "@every 1m",
        "overwrite": true
    }
  },
```

Using the official Pachyderm `pachctl` image, 
the transform first updates the default `pachctl` config
so `pachctl` can talk directly to `pachd` in the cluster.
It uses the `kubedns` name for `pachd` 
and the internal Service port of `650`. 

```
          "echo '{\"pachd_address\": \"grpc://pachd:650\"}' | pachctl config set context default --overwrite",
```

Similar to the deferred processing example,
the next command moves the `master` branch on `edges_dp` to point to `dev`,


```
          "pachctl create branch edges_dp@master --head dev"
```

This is all the cron pipeline needs to do, 
without access controls.
The `transform` section of the pipeline spec `branch-mover-no-auth.json`
will look like this:

```
  "transform": {
      "cmd": ["sh" ],
      "stdin": [
          "echo '{\"pachd_address\": \"grpc://pachd:650\"}' | pachctl config set context default --overwrite",
          "pachctl create branch edges_dp@master --head dev"
      ],
    "image": "pachyderm/pachctl:1.11.0"
  }
```

### Branch mover with access controls

Use the instructions in this section
if you have activated access controls in your Pachyderm cluster.
Otherwise, go back to  [Branch mover without access controls](#branch_mover_without_access_controls).

Adding support for access controls to the `branch-mover` pipeline requires a few steps.

1. Creating a [Kubernetes Secret](https://kubernetes.io/docs/concepts/configuration/secret/)
   containing an authentication token.
2. Loading that secret into Kubernetes using `pachctl create secret`.
3. Adding a `.transform.secret` to the pipeline spec 
   to create an environment variable from a key value in the secret.
4. Adding a line to the pipeline transform to authenticate using the token prior to moving the branch.

Let's go through each of these steps in detail.

#### Creating the authentication token and the secret

Once Pachyderm access controls are activated,
log in as the user the `branch-mover` 
will authenticate as to run this example.

You may want to test this with the `robot:admin`
configured when access controls were activated,
or your own credentials.
Please see [Using this example in production](#using_this_example_in_production) below
for information regarding production-level security configuration.

Create a Pachyderm authentication token by running the following command:

```
pachctl auth get-auth-token --ttl <some-golang-formatted-duration>
```

A golang-formatted duration uses `h` for hours, `m` for minutes, `s` for seconds.
26 weeks would be `24 * 7 * 26` hours, 
expressed as `624h`. 
The token will only be generated for this duration
if it is *shorter* than the lifetime of the session
for the user who is logged into the cluster
where the command is run. 
Otherwise, it is generated for the duration of that user's current session.
The expiration of a user's current session can be determined
by running `pachctl auth whomai`.

The duration of the token 
determines how long the cron pipeline may run 
before the secret needs to be refreshed 
and the pipeline restarted.

Here is a Unix command 
for generating a token using `pachctl`
and only outputting the value of the token:

```
pachctl auth get-auth-token --ttl "624h" | \
    grep Token | awk '{print $2}' | \
```

The command is enhanced to encode the token with the `base64` encoding scheme,
so it can be used in a Kubernetes secret,
and trim off unnecessary characters.

```
pachctl auth get-auth-token --ttl "624h" | \
    grep Token | awk '{print $2}' | \
    base64 -e | tr -d '\r\n'
```

Next, that data must be placed into a secret.
The template for an appropriate secret 
is in the file `pachyderm-user-secret.secret`.
The `jq` utility enables you to place the encoded token 
in the proper `data.auth_token` field
in the secret
by using a subshell to run that command
and direct the output into a json file,
which we'll give the `secret` extension.

```
jq ".data.auth_token=\"$(pachctl auth get-auth-token --ttl "624h" | \
    grep Token | awk '{print $2}' | \
    base64 -e | tr -d '\r\n')\"" \
    < pachyderm-user-secret.clear \
    > pachyderm-user-secret.secret
```

#### Loading the secret into Kubernetes

Next, let us load the secret into Kubernetes by running the following command:

```
pachctl create secret -f pachyderm-user-secret.secret
```

!!! note
    You can run the two previous steps by running
    `make pachyderm-user-secret.secret`.

#### Mounting the secret in the pipeline

To add the secret to our pipeline,
we can just use the `transform.secrets` field
to expose the `auth_token` key as an environment variable.
This is `transform.secrets` in the file `branch-mover.json`

```
      "secrets": [ {
          "name": "pachyderm-user-secret",
          "env_var": "PACHYDERM_AUTH_TOKEN",
          "key": "auth_token"
      } ]
```

#### Authenticating to Pachyderm

The `branch-mover.json` file includes one line
that uses the `PACHYDERM_AUTH_TOKEN` environment variable
to authenticate to Pachyderm. 

```
          "echo ${PACHYDERM_AUTH_TOKEN} | pachctl auth use-auth-token"
```

That line is inserted prior to creating the branch, 
making the pipeline transform in `branch-mover.json`
look like this:

```
"transform": {
      "cmd": ["sh" ],
      "stdin": [
          "echo '{\"pachd_address\": \"grpc://pachd:650\"}' | pachctl config set context default --overwrite",
          "echo ${PACHYDERM_AUTH_TOKEN} | pachctl auth use-auth-token",
          "pachctl create branch edges_dp@master --head dev"
      ],
    "image": "pachyderm/pachctl:1.11.0"
  }
```

#### Creating the pipeline

Finally, create the pipeline using that spec:

```
pachctl create pipeline -f branch-mover.json
```

!!! note
    You can run this step with the command  `make create-branch-mover`.

## Example run-through

This example can be used with access controls activated or not.
The only difference is the command that you use to create the pipeline 
in the second step, below.

1. If the DAG
   used by the deferred processing example
   hasn't yet been created,
   create that starting DAG 
   by running this command 
   from inside this directory.
   
   ```
   make create-deferred-processing-cluster
   ```

1. If your Pachyderm cluster does not have access controls activated,
   create the branch-mover cron pipeline
   using the `create-branch-mover-no-auth` Makefile target.
   
   ```
   make create-branch-mover-no-auth
   ```
   
   If you have access controls activated,
   create the branch-mover cron pipeline
   using the `create-branch-mover` Makefile target.
   
   ```
   make create-branch-mover
   ```

1. Watch `pachctl jobs` in another terminal window
   by using this command:
   
   ```
   watch -cn 2 pachctl list job --no-pager
   ```

!!! note
    On macOS, you may need to install `watch`, 
        which may be installed via [Homebrew](https://brew.sh/)
        using the command `brew install watch`.

1. Every minute, you should see a job triggered on `branch-mover`.
   The very first job will be immediately followed
   by a job for `montage_dp`, 
   as existing files are moved to the `edges_dp@master` branch.
   Subsequent ticks will trigger no jobs in `montage_dp`.

1. Commit data to the `images_dp_1` repo.

```
pachctl put file images_dp_1@master:1VqcWw9.jpg -f http://imgur.com/1VqcWw9.jpg
```

   A job will be triggered on `edges_dp`, 
   but no jobs will be triggered on `montage_dp`
   until after `branch-mover` runs
   moving the `edges@dev` branch to `edges@master`.

## Using this example in production

When you implement this example on production pipelines with access controls activated,
you will have to periodically renew the token
by either running the appropriate make target
to update the pipeline with a new secret
or manually updating the secret
and deleting and recreating the pipeline.

It is a best security practice in production
to create a Pachyderm user 
with the [least privilege](https://en.wikipedia.org/wiki/Principle_of_least_privilege) required to do this pipeline's tasks.

This is a periodic maintenance task
with security implications
the automation of which should be reviewed
by appropropriate engineering personnel.

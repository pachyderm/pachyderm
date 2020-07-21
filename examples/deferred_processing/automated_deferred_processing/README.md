# Automated Deferred Processing 

[Deferred processing](https://docs.pachyderm.com/latest/how-tos/deferred_processing/) 
is a Pachyderm technique for controlling when data gets processed.
Deferred processing uses branches to prevent pipelines from triggering on every input commit.
This example shows how to automate the movement of those branches.

The example here triggers a job periodically, using a cron pipeline.
The Makefile you'll find in this example,
along with the explanations provided in this document,
should give you a good start on implementing this in your Pachyderm cluster
with or without authentication enabled.

## Prerequisites

Before you begin, you should understand deferred processing by reading the documentation
and trying the [deferred processing example](../deferred_processing_plus_transactions).
That example is used extensively here.

If you're using a Pachyderm cluster with authentication activated,
this example will show you how to create a Pachyderm authentication token,
load the token into a Kubernetes secret provisioned through `pachctl`,
and use `pod_patch` to mount the secret as a Kubernetes volume
and create environment variables for use by the pipeline.
If you are unfamiliar with those things,
it may be helpful to refer to the following documentation as you work through the example.
* [Pachyderm authentication documentation](https://docs.pachyderm.com/latest/enterprise/auth/)
* [Kubernetes documentation on
Secrets](https://kubernetes.io/docs/concepts/configuration/secret/),
* [pachctl create secret](https://docs.pachyderm.com/latest/reference/pachctl/pachctl_create_secret/) command
* [Pachyderm documentation on adding a volume to a pipeline](https://docs.pachyderm.com/latest/how-tos/mount-volume/) using [pod_patch](https://docs.pachyderm.com/latest/reference/pipeline_spec/#pod-patch-optional)
* the [jq utility](https://stedolan.github.io/jq/manual/) for transforming json files in shell scripts

You need to have Pachyderm v1.9.8 or later installed on your computer or cloud platform. 
See [Deploy Pachyderm](https://docs.pachyderm.com/latest/deploy-manage/deploy/).

It helps to have some understanding of Makefiles, 
if you plan on using those in your infrastructure.
Basic Unix shell scripting skills are assumed.

## Pipelines

This example uses the same DAG as in the deferred processing example,
with the addition of a cron pipeline 
for periodically moving the `dev` branch to `master`.

For details on the deferred processing example DAG,
please refer to [that example](../deferred_processing_plus_transactions).

### Branch mover, without authentication

The cron pipeline is called `branch-mover`. 
It is configured to run, by default, every 1 minute, 
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

The code for moving the branch is very simple, 
if the Pachyderm cluster does not have authentication activated.

Using the official Pachyderm pachctl image, 
the transform first updates the default `pachctl` config
so `pachctl` can talk directly to `pachd` in the cluster.
It uses the `kubedns` name for `pachd`, 
and the internal Service port of `650`. 

```
          "echo '{\"pachd_address\": \"grpc://pachd:650\"}' | pachctl config set context default --overwrite",
```

The next command moves the `master` branch on `edges_dp` to point to `dev`,
just like in the deferred processing example.

```
          "pachctl create branch edges_dp@master --head dev"
```

This is all the cron pipeline needs to do, 
without authentication enabled. 
The transform in the pipeline spec `branch-mover-no-auth.pipeline`
ends up looking like this:

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

### Branch mover, with authentication

Adding authentication to the `branch-mover` pipeline requires a few steps.

1. Creating a Kubernetes Secret containing an authentication token.
1. Loading that secret into Kubernetes using `pachctl create secret`.
1. Adding a [pod_patch](https://docs.pachyderm.com/latest/reference/pipeline_spec/#pod-patch-optional) to mount the secret as a volume in the pipeline and create environment variables
1. Adding a line to the pipeline transform to authenticate using the token prior to moving the branch.

Let's go through each of these steps in detail.

#### Creating the authentication token and the secret

Once Pachyderm authentication is activated,
log in as the user the branch-mover 
will authenticate as to run this example.

You may want to first test with the `robot:admin`
configured when authentication was set up,
or your own credentials,
but it is a best security practice in production
to create a Pachyderm user 
with the [least privilege](https://en.wikipedia.org/wiki/Principle_of_least_privilege) required to do the task this pipeline requires.

A Pachyderm authentication token is created with the command
```
pachctl auth get-auth-token --ttl <some-golang-formatted-duration>
```
A golang-formatted duration uses `h` for hours, `m` for minutes, `s` for seconds.
26 weeks would be `24 * 7 * 26` hours, 
expressed as `624h`. 
The token will only be generated for this duration
if it is *shorter* than the lifetime of the session
for the user who's logged into the cluster
where the command is run. 
Otherwise, it's generated the lifetime of the user's current session.
The expiration of the current session can be determined
by running `pachctl auth whomai`.

The duration of the token 
determines how long the cron pipeline may run 
before the secret needs to be refreshed 
and the pipeline restarted.

Here is a Unix command 
for generating a token using `pachctl`
and only outputting the value of the token

```
pachctl auth get-auth-token --ttl "624h" | \
    grep Token | awk '{print $2}' | \
```

The command is enhanced to base64 encode the token,
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
The jq utility allows the encoded token 
to be placed in the proper `data.auth_token` field
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

Next, that secret must be loaded into Kubernetes.
```
pachctl create secret -f pachyderm-user-secret.secret
```

.NOTE
The prior two steps are in the Makefile
in the  `pachyderm-user-secret.secret` target.

#### Mounting the secret in the pipeline

The template for the `pod_patch`
that mounts the secret in a volume
and makes the token available as an environment variable
is in the file `json-patch.template`.
It was created using the process described
in the Pachyderm documentation on adding a volume to a pipeline. 

The Unix command below will use the Unix `sed` and `tr` utilities
to format the patch for best practices.
It will also remove newlines 
and escape `"` characters,
required to include it in pipeline specification.
The sed commands themselves are in `commands.sed`.
The json patch will be output in JSON 
to a file with a `jpatch` extension.

```
sed -E -f commands.sed json-patch.template | \
    tr '\n' ' ' > json-patch.jpatch
```

.NOTE
The prior step is in the Makefile
in the  `json-patch.jpatch` target.

#### Adding the patch and using it in the pipeline
To add the json patch to our pipeline,
we can just paste it in,
or run the Unix command below to add it
to an otherwise valid JSON pipeline spec 
in the file `branch-mover.pipeline`

```
jq ".pod_patch=\"$(cat json-patch.jpatch)\"" branch-mover.pipeline \
        > branch-mover.json
```

That file, `branch-mover.pipeline`, includes one line
that uses the `PACHYDERM_AUTH_TOKEN` environment variable
to authenticate to Pachyderm. 

```
          "echo ${PACHYDERM_AUTH_TOKEN} | pachctl auth use-auth-token"
```

That line is inserted prior to creating the branch, 
making the pipeline transform in `branch-mover.pipeline`
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

.NOTE
The prior step is in the Makefile
in the  `branch-mover.json` target.

#### Create the pipeline

Finally, create the pipeline using that spec

```
pachctl create pipeline -f branch-mover.json
```

.NOTE
The prior step is in the Makefile
in the  `create-branch-mover` target.

## Example run-through

This example can be used with no authentication or with authentication activated.
The only difference is the command you use to create the pipeline 
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

2. If you are not using authentication,
   create the branch-mover cron pipeline
   using the `create-branch-mover-no-auth` makefile target.
   
   ```
   make create-branch-mover-no-auth
   ```

   If you have authentication activated,
   create the branch-mover cron pipeline
   using the `create-branch-mover` makefile target.
   
   ```
   make create-branch-mover
   ```

3. Watch pachctl jobs in another terminal window
   using this command. 
   
   ```
   watch -cn 2 pachctl list job --no-pager
   ```

.NOTE On macOS, you may need to install `watch`, 
which may be installed via [Homebrew](https://brew.sh/)
using the command `brew install watch`
   
4. Every minute, you should see a job triggered on `branch-mover`.


5. Commit data to the `images_dp_1` repo.

```
pachctl put file images_dp_1@master:1VqcWw9.jpg -f http://imgur.com/1VqcWw9.jpg
```

6. A job will be triggered on `edges_dp`, 
but no jobs will be triggered on `montage_dp`
until after `branch-mover` runs
moving the `edges@dev` branch to `edges@master`.

## Notes on using this in production

When used on production pipelines with authentication enabled,
you will have to periodically renew the token
by either running the appropriate make target
to update the pipeline with a new secret
or manually updating the secret
and deleting and recreating the pipeline.

This is a periodic maintenance task
with security implications
the automation of which should be reviewed
by appropropriate engineering personnel.





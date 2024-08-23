# Running Pachyderm locally

If you've read [BAZEL.md](BAZEL.md), you can edit and build the code. Next, you'll want to run it.
There are two ways to do this.

## testpachd

testpachd is most of pachd, but without any ability to actually run pipelines. To start it up, run

    bazel run //src/testing/testpachd

This will automatically edit your Pachyderm config file to point at this daemon (and restore it when
the daemon exits; press Control-C). You can then run pretty much any pachctl commands you desire.

There are some flags you can pass:

- `-log`: Log to a file instead of the default. (The default logger creates a log file at DEBUG
  level in `$XDG_CACHE_DIR` that is deleted if testpachd exits without error. The usual path is
  someting like `~/.cache/pachyderm/log/testpachd.20240326T015415Z.log`. The timestamp is in UTC.)
- `-v`: Print DEBUG level logs.
- `-auth`: Activate auth. (Your pach context will be set to log in as root, but you can create a
  user and log in with that if you want.)
- `-loki`: Run a Loki server alongside pachd, and send pachd logs to it. This is useful for quickly
  testing `pachctl logs`.

As always, you need to hide these flags from Bazel, like:
`bazel run //src/testing/testpachd -- -auth`.

Docker is required, as `dockertestenv` is used to run postgres and minio.

## pachdev

If you need to install Pachyderm for real, with console and Kubernetes and all that good stuff,
you'll want to use pachdev. All dependencies are bundled with this repo, so you don't need to
install anything other than Bazel.

### Create a k8s cluster

First, create a new k8s cluster:

    bazel run //src/testing/pachdev create-cluster -- [<optional name>] <flags>

This is where you set up some important configuration:

- `<optional name>`: You can have more than one cluster on your machine at a time. This will be the
  name of it. If you don't set a name, `pach` will be used. The k8s context
  (`kubectl --context=...`) will be called `kind-pach` and the pach context will be called `pach`.
  If you set a name, it will `kind-pach-<name>` and `<name>`.
- `--hostname=<string>`: Set the public hostname for this instance. If you're running on your VM and
  want to access console in your browser, set this to the external IP of the VM. `localhost` doesn't
  work well; prefer `127.0.0.1` if that's what you want. (You'll need to type http://127.0.0.1 into
  your browser as well.)
- `--http=<bool>`: Defaults to true. If true, this cluster will take over ports 80 and 443 on your
  machine, for web access to Pachyderm. Only one cluster can do this, so when you create a second
  cluster, you'll have to set `--http=false` or the cluster will fail to create.
- `--push-path=<string>`: If set, we'll push containers to this Skopeo URL. If unset, we'll
  automatically start a Docker registry on your local machine and use our fast copy hacks to get
  images into the cluster. You pretty much never want to set this unless you Know What You're Doing.
- `--test-namespaces=<int>`: How many extra "test namespaces" to create; this bounds the number of
  k8s tests that can run concurrently.
- `--starting-port=<uint16>`: If set, we'll allocate `10 * (<number of test namespaces> + 1)` ports
  for tests that install non-Pachyderm services to k8s but need to access those services over the
  network. If -1, don't bother allocating these ports. -1 is a reasonable default, but if you pass
  `--http=false`, you'll need to set an explicit port for pachd (or `0` and we'll pick one).

`--hostname` is typically all you need to set. The default isn't too useful and might not work with
your setup.

This command requires docker, since the k8s cluster is run with kind, Kubernetes IN Docker. You can
try podman but you'll have to adjust some flags.

### Push your code

Once you have a functional cluster, push your code with:

    bazel run //src/testing/pachdev push

This will rebuild pachd, the worker, etc. push the images to your cluster, run helm (against
//etc/helm/pachyderm in your working copy), wait for the new version to be ready, and adjust
`pachctl` to talk to this cluster.

No error messages are generated if your pachd crashes at startup, it will just hang forever and
you'll have to investigate with `kubectl` (but a future `push` with non-broken code will fix the
error state). The first time you run `push`, it's actually a fresh installation and it will take a
few minutes to fetch dependencies like loki, postgres, etcd, etc. The next time you run `push`, it's
typically very fast, on the order of 10 seconds after the build completes and `pachdev` actually
starts running.

Running `--help` describes some options that may be of use. Protect the `--help` flag from
`bazel run` as you always do.

### Delete a cluster

    bazel run //src/testing/pachdev delete-cluster [<optional name>]

This will completely remove any trace of the named (or current, based on k8s context in
`~/.kube/config`) cluster. This is mostly necessary when you need to reset Postgres because you
added a bad migration or something. Use `testpachd` to test your database migrations and save
yourself the trouble .

### Maintenance

We leave some state around that might fill up your disk. All built containers are stored in
`${XDG_STATE_HOME}/<cluster name>-registry/registry`. This is typically something like
`~/.local/state/pach-registry/registry/` for an unnamed cluster.

The state is kept so that you can rollback `pachdev push` with helm.

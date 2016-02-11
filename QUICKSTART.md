# Quick Start Guide: A Grep Pipeline

In this guide you're going to create a Pachyderm pipeline using one of the unix
classics, `grep`.  For those unfamiliar, `grep` lets you search files with
regexes.  Our `grep` pipeline will allow us to scale `grep`'s regex search up
to a virtually limitless stream of data. Let's dive in.

## Setup

Before we can launch a cluster you'll need the following things:

- Docker >= 1.9 (must deploy with [`--storage-driver=devicemapper`](http://muehe.org/posts/switching-docker-from-aufs-to-devicemapper/))
- Go >= 1.5
- Kubernetes and Kubectl >= 1.1.2
- FUSE 2.8.2 (https://osxfuse.github.io/)

### Kubernetes

For development we recommend a dockerized Kubernetes deployment.
If `docker` is installed you can launch one from the root of this repo with:

```shell
$ make kube-launch
```

### Forward Ports
Unless `docker` is running locally we'll need to forward ports so that `kubectl`
and `pachctl` can talk to the servers.

```shell
$ ssh <HOST> -fTNL 8080:localhost:8080 -L 30650:localhost:30650
```

### `pachctl`
Pachyderm is controlled with a CLI, `pachctl`. To install:

```shell
$ make install
```

## Launch the Cluster

Launching a Pachyderm cluster from the root of this repo is dead simple:

```shell
$ make launch
```

This passes the manifest located at `etc/kube/pachyderm.json` to `kubectl create`.
Here's what a functioning cluster looks like:

```shell
$ kubectl get all
CONTROLLER   CONTAINER(S)   IMAGE(S)                               SELECTOR      REPLICAS   AGE
etcd         etcd           gcr.io/google_containers/etcd:2.0.12   app=etcd      1          2m
pachd        pachd          pachyderm/pachd                        app=pachd     1          1m
rethink      rethink        rethinkdb:2.1.5                        app=rethink   1          1m
NAME         CLUSTER_IP   EXTERNAL_IP   PORT(S)                        SELECTOR      AGE
etcd         10.0.0.197   <none>        2379/TCP,2380/TCP              app=etcd      2m
kubernetes   10.0.0.1     <none>        443/TCP                        <none>        10d
pachd        10.0.0.100   nodes         650/TCP,750/TCP                app=pachd     1m
rethink      10.0.0.218   <none>        8080/TCP,28015/TCP,29015/TCP   app=rethink   1m
NAME                   READY     STATUS    RESTARTS   AGE
etcd-4r4hp             1/1       Running   0          2m
k8s-master-127.0.0.1   3/3       Running   0          10d
pachd-u992h            1/1       Running   1          1m
rethink-268hq          1/1       Running   0          1m
NAME      LABELS    STATUS    VOLUME    CAPACITY   ACCESSMODES   AGE
```

## Mount the Filesystem
The first thing we need to do is mount Pachyderm's filesystem (`pfs`) so that we
can read and write data.
```shell
# We background this process because it blocks.
$ pachctl mount &
```

This will mount pfs on `/pfs` you can inspect the filesystem like you would any
other local filesystem. Try:

```shell
$ ls /pfs
```
That probably wasn't terribly interesting, that's ok you shouldn't see anything
yet, `/pfs` will contain a directory for each `repo`, but you haven't made any
yet. Let's make one.

## Create a `Repo`

`Repo`s are the highest level primitive in `pfs`, like all primitives in pfs they share
their name with a primitive in `git` and are designed to behave analagously.
Generally `repo`s should be dedicated to a single source of data, for example log
messages from a particular service. `Repo`s are dirt cheap so don't be shy about
making them very specific. For this demo we'll simply create a `repo` called
"data" to hold the data we want to grep:

```shell
$ pachctl create-repo data
$ ls /pfs
data
```

Now `ls` does something! `/pfs` contains a directory for every repo in the
filesyste.

## Create a `Commit`
Now that you've created a `Repo` you should see an empy directory `/pfs/data`
if you try writing to it, it will fail because you can't write directly to
`Repo`s. In Pachyderm you write to explicit commits, let's start a new commit:

```shell
$ pachctl start-commit data
6a7ddaf3704b4cb6ae4ec73522efe05f
```

this returns a brand new commit id, yours should be different from mine.
Now if we take a look back at `/pfs` things have changed:

```shell
$ ls /pfs/data
6a7ddaf3704b4cb6ae4ec73522efe05f
```

a new directory has been created for our commit. This we can write to.

```shell
$ echo foo >/pfs/data/6a7ddaf3704b4cb6ae4ec73522efe05f/file
```

However if you try to view the file you'll notice you can't:

```shell
$ cat /pfs/data/6a7ddaf3704b4cb6ae4ec73522efe05f/file
cat: /pfs/data/6a7ddaf3704b4cb6ae4ec73522efe05f/file: No such file or directory
```

Pachyderm won't let you read data from a commit until the commit is finished.
This prevents reads from racing with other writes and makes every write to the
filesystem atomic. Let's finish the commit:

```shell
$ pachctl finish-commit data 6a7ddaf3704b4cb6ae4ec73522efe05f
```

Now we can view the files:

```shell
$ cat /pfs/data/6a7ddaf3704b4cb6ae4ec73522efe05f/file
foo
```

However, we've lost the ability to write to it, finished commits are immutable.

## Create a `Pipeline`

Now that we've got some data in our `repo` it's time to do something with it.
`Pipeline`s are an increasingly common paradigm for distributed processing, the
essence of the paradigm is that processing elements are chained together with
the output of one feeding into the input of the next. Pipelines are the core
primitive for Pachyderm's processing system (pps), they're specified with a JSON
encoding. The `pipeline` we're creating can be found at
`examples/grep/pipeline.json` here's what it looks like:

```json
{
  "pipeline": {
    "name": "grep-example"
  },
  "transform": {
    "cmd": [ "sh" ],
    "stdin": "grep -r \"foo\" /pfs/data > /pfs/out/foo"
  },
  "shards": "1",
  "inputs": [ { "repo": { "name": "data" } } ]
}
```

We can create it with:

```shell
$ pachctl create-pipeline -f examples/grep/pipeline.json
```

## What Happens When You Create a Pipeline
Creating a `pipeline` tells Pachyderm to run your code on *every* finished
`commit` in a `repo`, including `commit`s that happen after the `pipeline` is
created. Our `repo` already had a `commit` so Pachyderm will have already
launched a `job` to process that `commit`.

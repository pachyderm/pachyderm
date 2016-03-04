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
That probably wasn't terribly interesting, but that's ok because you shouldn't see anything
yet. `/pfs` will contain a directory for each `repo`, but you haven't made any
yet. Let's make one.

## Create a `Repo`

`Repo`s are the highest level primitive in `pfs`. Like all primitives in pfs, they share
their name with a primitive in Git and are designed to behave analagously.
Generally, `repo`s should be dedicated to a single source of data, for example log
messages from a particular service. `Repo`s are dirt cheap so don't be shy about
making them very specific. For this demo we'll simply create a `repo` called
"data" to hold the data we want to grep:

```shell
$ pachctl create-repo data
$ ls /pfs
data
```

Now `ls` does something! `/pfs` contains a directory for every repo in the
filesystem.

## Start a `Commit`
Now that you've created a `Repo` you should see an empty directory `/pfs/data`.
If you try writing to it, it will fail because you can't write directly to a
`Repo`. In Pachyderm, you write data to an explicit `commit`. Commits are
immutable snapshots of your data which give Pachyderm its version control for
data properties. Unlike Git though, commits in Pachyderm must be explicitly
started and finished.

Let's start a new commit:
```shell
$ pachctl start-commit data
6a7ddaf3704b4cb6ae4ec73522efe05f
```

This returns a brand new commit id. Yours should be different from mine.
Now if we take a look back at `/pfs` things have changed:
```shell
$ ls /pfs/data
6a7ddaf3704b4cb6ae4ec73522efe05f
```

A new directory has been created for our commit and now we can start adding
files. We've provided some sample data for you to use -- a list of purchases
from a fruit stand. We're going to write that data as a file "sales" in pfs.

```shell
# Write sample data to pfs
$ cat examples/grep/set1.txt >/pfs/data/6a7ddaf3704b4cb6ae4ec73522efe05f/sales
```

However, you'll notice that we can't read the file "sales" yet.

```shell
$ cat /pfs/data/6a7ddaf3704b4cb6ae4ec73522efe05f/sales
cat: /pfs/data/6a7ddaf3704b4cb6ae4ec73522efe05f/sales: No such file or directory
```

## Finish a `Commit`

Pachyderm won't let you read data from a commit until the `commit` is `finished`.
This prevents reads from racing with writes. Furthermore, every write
to pfs is atomic. Now let's finish the commit:

```shell
$ pachctl finish-commit data 6a7ddaf3704b4cb6ae4ec73522efe05f
```

Now we can view the file:

```shell
$ cat /pfs/data/6a7ddaf3704b4cb6ae4ec73522efe05f/sales
```
However, we've lost the ability to write to this `commit` since finished
commits are immutable. In Pachyderm, a `commit` is always either _write-only_
when it's been started and files are being added, or _read-only_ after it's
finished. 


## Create a `Pipeline`

Now that we've got some data in our `repo` it's time to do something with it.
Pipelines are the core primitive for Pachyderm's processing system (pps) and
they're specified with a JSON encoding. The `pipeline` we're creating
can be found at `examples/grep/pipeline.json`. Here's what it looks
like:

```json

{
  "pipeline": {
    "name": "grep"
  },
  "transform": {
    "cmd": [ "sh" ],
    "stdin": [
        "grep -r apple /pfs/data >/pfs/out/apple",
        "grep -r banana /pfs/data >/pfs/out/banana",
        "grep -r orange /pfs/data >/pfs/out/orange"
    ]
  },
  "shards": "1",
  "inputs": [{"repo": {"name": "data"}}]
}
```
In this `pipeline`, we are grepping for the terms "apple", "orange", and
"banana" and writing that line to the corresponding file. Notice we read data
from `/pfs/` and write data to `/pfs/out/`. In this example, the output of our
pipeline is three files, one for each type of fruit sold with a list of all
purchases of that fruit. 

Now we create the grep pipeline in Pachyderm:
```shell
$ pachctl create-pipeline -f examples/grep/pipeline.json
```

## What Happens When You Create a Pipeline
Creating a `pipeline` tells Pachyderm to run your code on *every* finished
`commit` in a `repo` as well as all future commits that happen after the pipeline is
created. Our `repo` already had a `commit` so Pachyderm will automatically
launch a `job` to process that data.

You can view the job with:

```shell
$ pachctl list-job
ID                                 OUTPUT                                  STATE
09a7eb68995c43979cba2b0d29432073   grep/2b43def9b52b4fdfadd95a70215e90c9   JOB_STATE_RUNNING
```

Depending on how quickly you do the above, you may see `JOB_STATE_RUNNING` or
`JOB_STATE_SUCCESS` (hopefully you won't see `JOB_STATE_FAILURE`).

Pachyderm `job`s are implemented as Kubernetes jobs, so you can also see your job with:

```shell
$ kubectl get job
JOB                                CONTAINER(S)   IMAGE(S)             SELECTOR                                                         SUCCESSFUL
09a7eb68995c43979cba2b0d29432073   user           pachyderm/job-shim   app in (09a7eb68995c43979cba2b0d29432073),suite in (pachyderm)   1
```

## Reading the Output

Every `pipeline` outputs its results to a corresponding `repo` with the same
name. In our example, the "grep" pipeline created a repo "grep" where it stored
the output files. You can read the out data the same way that we read the input
data:

```shell
$ cat /pfs/grep/2b43def9b52b4fdfadd95a70215e90c9/apple
```

## Processing More Data

Pipelines will also automatically process the data from new commits as they are
created. Think of pipelines as being subscribed to any new commits that are
finished on their input repo(s). Also similar to Git, commits have a parental
structure that track how files change over time. Specifying a parent is
optional when creating a commit (notice we didn't specify a parent when we
created the first commit), but in this case we're going to be adding
more data to the same file "sales."

Let's create a new commit with our previous commit as the parent:

```shell
$ pachctl start-commit data 6a7ddaf3704b4cb6ae4ec73522efe05f
fab8c59c786842ccaf20589e15606604
```

Next, we need to add more data. We're going to append more purchases from set2.txt to the file "sales."

```shell
$ cat examples/grep/set2.txt >/pfs/data/fab8c59c786842ccaf20589e15606604/sales
```
Finally, we'll want to finish our second commit. After it's finished, we can
read "sales" from the latest commit to see all the puchases from `set1` and
`set2`. We could also chose to read from the first commit to only see `set1`.

```shell
$ pachctl finish-commit data fab8c59c786842ccaf20589e15606604
```
Finishing this commit will automatically trigger the grep pipeline to run on
the new data we've added. We'll also see a corresponding comit to the output
"grep" repo with appends the the "apple", "orange" and "banana" files.

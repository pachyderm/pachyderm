# Quick Start Guide: Fruit Stand

In this guide you're going to create a Pachyderm pipeline to process
transaction logs from a fruit stand. We'll use two standard unix tools, `grep`
and `awk` to do our processing. Thanks to Pachyderm's processing system we'll
be able to run the pipeline in a distributed, streaming fashion. As new data is
added the pipeline will automatically process it and materialize the results.

## Setup

Before we can launch a cluster you'll need the following things:

- Docker >= 1.9 (must deploy with [`--storage-driver=devicemapper`](http://muehe.org/posts/switching-docker-from-aufs-to-devicemapper/))
- Go >= 1.5
- Kubernetes and Kubectl >= 1.2.0
- FUSE 2.8.2 (https://osxfuse.github.io/)

### Kubernetes

For development we recommend a dockerized Kubernetes deployment.
If `docker` is installed you can launch one from the root of this repo with:

```shell
$ make launch-kube
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
$ kubectl create -f http://pachyderm.io/manifest.json
```

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
"data" to hold the data we want to process:

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
$ cat examples/fruit_stand/set1.txt >/pfs/data/6a7ddaf3704b4cb6ae4ec73522efe05f/sales
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
they're specified with a JSON encoding. We're going to create a pipeline with 2
transformations in it. The first transformation filters the sales logs into separate records for apples,
oranges and bananas using `grep`. The second one uses `awk` to sum these sales numbers into a final sales count.

```
+----------+     +------------+     +------------+
|input data| --> |filter pipline| --> |sum pipeline|
+----------+     +------------+     +------------+
```

The `pipeline` we're creating can be found at `examples/fruit_stand/pipeline.json`.
Here's what it looks like:
```json
{
  "pipeline": {
    "name": "filter"
  },
  "transform": {
    "cmd": [ "sh" ],
    "stdin": [
        "grep apple  /pfs/data/sales >/pfs/out/apple",
        "grep banana /pfs/data/sales >/pfs/out/banana",
        "grep orange /pfs/data/sales >/pfs/out/orange"
    ]
  },
  "shards": "1",
  "inputs": [
    {
      "repo": {
        "name": "data"
      }
    }
  ]
}
{
  "pipeline": {
    "name": "sum"
  },
  "transform": {
    "cmd": [ "sh" ],
    "stdin": [
        "cut -f 2 /pfs/filter/apple | awk '{s+=$1} END {print s}' >/pfs/out/apple",
        "cut -f 2 /pfs/filter/banana | awk '{s+=$1} END {print s}' >/pfs/out/banana",
        "cut -f 2 /pfs/filter/orange | awk '{s+=$1} END {print s}' >/pfs/out/orange"
    ]
  },
  "shards": "1",
  "inputs": [
    {
      "repo": {
        "name": "filter"
      },
	  "reduce": true
    }
  ]
}
```

In the first step of this `pipeline`, we are grepping for the terms "apple", "orange", and
"banana" and writing that line to the corresponding file. Notice we read data
from `/pfs/data` (/pfs/[input_repo_name]) and write data to `/pfs/out/`. The second step of this pipeline takes each file, removes the fruit name, and sums up the purchases. The output of our complete
pipeline is three files, one for each type of fruit with a single number showing the total quantity sold. 

Now let's create the pipeline in Pachyderm:

```shell
$ pachctl create-pipeline -f examples/fruit_stand/pipeline.json
```

## What Happens When You Create a Pipeline
Creating a `pipeline` tells Pachyderm to run your code on *every* finished
`commit` in a `repo` as well as *all future commits* that happen after the pipeline is
created. Our `repo` already had a `commit` so Pachyderm will automatically
launch a `job` to process that data.

You can view the job with:

```shell
$ pachctl list-job
ID                                 OUTPUT                                  STATE
09a7eb68995c43979cba2b0d29432073   filter/2b43def9b52b4fdfadd95a70215e90c9   JOB_STATE_RUNNING
```

Depending on how quickly you do the above, you may see `JOB_STATE_RUNNING` or
`JOB_STATE_SUCCESS` (hopefully you won't see `JOB_STATE_FAILURE`).

Pachyderm `job`s are implemented as Kubernetes jobs, so you can also see your job with:

```shell
$ kubectl get job
JOB                                CONTAINER(S)   IMAGE(S)             SELECTOR                                                         SUCCESSFUL
09a7eb68995c43979cba2b0d29432073   user           pachyderm/job-shim   app in (09a7eb68995c43979cba2b0d29432073),suite in (pachyderm)   1
```

Every `pipeline` creates a corresponding `repo` with the same
name where it stores its output results. In our example, the "filter" transformation created a `repo` called "filter" which was the input to the "sum" transformation. The "sum" `repo` contains the final output files.

## Reading the Output
 We can read the output data from the "sum" `repo` in the same fashion that we read the input data:

```shell
$ cat /pfs/sum/2b43def9b52b4fdfadd95a70215e90c9/apple
```

## Processing More Data

Pipelines will also automatically process the data from new commits as they are
created. Think of pipelines as being subscribed to any new commits that are
finished on their input repo(s). Also similar to Git, commits have a parental
structure that track how files change over time. Specifying a parent is
optional when creating a commit (notice we didn't specify a parent when we
created the first commit), but in this case we're going to be adding
more data to the same file "sales."

In our fruit stand example, this could be making a commit every hour with all the new purchases that happened in that timeframe. 

Let's create a new commit with our previous commit as the parent:

```shell
$ pachctl start-commit data -p 6a7ddaf3704b4cb6ae4ec73522efe05f
fab8c59c786842ccaf20589e15606604
```

Next, we need to add more data. We're going to append more purchases from set2.txt to the file "sales."

```shell
$ cat examples/fruit_stand/set2.txt >/pfs/data/fab8c59c786842ccaf20589e15606604/sales
```
Finally, we'll want to finish our second commit. After it's finished, we can
read "sales" from the latest commit to see all the puchases from `set1` and
`set2`. We could also chose to read from the first commit to only see `set1`.

```shell
$ pachctl finish-commit data fab8c59c786842ccaf20589e15606604
```
Finishing this commit will also automatically trigger the pipeline to run on
the new data we've added. We'll see a corresponding commit to the output
"sum" repo with files "apple", "orange" and "banana" each containing the cumulative total of purchases. Let's read the "apples" file again and see the new total number of apples sold. 

```shell
$ cat /pfs/sum/2b43def9b52b4fdfadd95a70215e90c9/apple
```
One thing that's interesting to note is that the first step in our pipeline is completely incremental. Since `grep` is a command that is completely parallelizable (i.e. it's a `map`), Pachyderm will only `grep` the new data from set2.txt. If you look back at the pipeline, you'll notice that there is a `"reduce": true` flag for "sum", which is an aggregation and is not done incrementally. Although many reduce operations could be computed incrementally, including sum, Pachyderm makes the safe choice to not do it by default. 

## Next Steps
You've now got a working Pachyderm cluster with data and a pipelines! You can continue to generate more data and commits and the Fruit Stand pipeline with automatically run to completion. Here are a few ideas for next steps that you can expand on your working setup. 

- Add a new pipeline that does something interesting with the "sum" repo as an input.
- Add your own data set and `grep` for different terms. This example can be generalized to generic word count. 
- If you're really feeling ambitious, you can create a much more complex pipeline that takes in any generic text and does some simple NLP on it. 

We'd love to help and see what you come up with so submit any issues/questions you come across or email at info@pachyderm.io if you want to show off anything nifty you've created! 


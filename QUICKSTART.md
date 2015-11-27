# Quick Start

### Prerequisites

- Docker >= 1.9
- Go >= 1.5
- Kubernetes >= 1.1.2

We're going to assume you have Docker and Go installed already, if not go do
that, both are pretty easy. Kubernetes can be a bit more finicky, if you've
already got a Kubernetes cluster you want to use (and `kubectl` works) then
great! Proceed to the next step. Otherwise run the following command from the
root of this repo to launch a local Kubernetes cluster on Docker:

```shell
$ etc/kube/start-kube-docker.sh
```

if Docker is running on another machine (such as via docker-machine) you'll
need to forward port 8080 to it:

```shell
$ ssh KUBEHOST -fTNL 8080:localhost:8080
```

### Install `pachctl`

From the root of this repo run:

```shell
$ make install
```

### Deploy a Cluster
```shell
$ pachctl create-cluster
```

This should create a new cluster, to check if it worked do:

```shell
$ kubectl get svc
NAME         CLUSTER_IP   EXTERNAL_IP   PORT(S)                        SELECTOR      AGE
etcd         10.0.0.5     <none>        2379/TCP,2380/TCP              app=etcd      7m
kubernetes   10.0.0.1     <none>        443/TCP                        <none>        3h
pfsd         10.0.0.65    <none>        650/TCP,750/TCP                app=pfsd      6m
ppsd         10.0.0.154   <none>        651/TCP                        app=ppsd      6m
rethink      10.0.0.186   <none>        8080/TCP,28015/TCP,29015/TCP   app=rethink   6m
```

### Forward Ports

As above, if Docker is running on another machine you'll need to forward some ports before you can access Pachyderm:

```shell
$ ssh KUBEHOST -fTNL 650:localhost:650
$ ssh KUBEHOST -fTNL 651:localhost:651
```

### Mount /pfs

Now we have a working Pachyderm cluster, let's do something with it.
Start by mounting the filesystem:

```shell
# We background this process because it blocks.
$ pachctl mount &
```

This will mount pfs on `/pfs` you can inspect the filesystem like you would any
other local filesystem. Try:

```shell
$ ls /pfs
```

You shouldn't see anything yet, `/pfs` will contain a directory for each
`Repo`, but you haven't made any yet. Let's make one.

### Creating a `Repo`

```shell
$ pachctl create-repo foo
$ ls /pfs
foo
```

### Creating a `Commit`
Now that you've created a `Repo` you should see an empty directory `/pfs/foo` if
you try writing to it, it will fail:

```shell
$ echo data >/pfs/foo/file
<figure out what this error message is>
```

That's because you can't write directly to `Repos`, you have to create a
`Commit` to write to instead:

```shell
$ pachctl start-commit foo
6a7ddaf3704b4cb6ae4ec73522efe05f
```

this returns a brand new commit id, yours should be different from mine.
Now if we take a look back at `/pfs` things have changed:

```shell
$ ls /pfs/foo
6a7ddaf3704b4cb6ae4ec73522efe05f
```

a new directory has been created for our commit. This we can write to:

```shell
$ echo foo >/pfs/foo/6a7ddaf3704b4cb6ae4ec73522efe05f/file
```

However if you try to view the file you'll notice you can't:

```shell
$ cat /pfs/foo/6a7ddaf3704b4cb6ae4ec73522efe05f/file
cat: /pfs/foo/6a7ddaf3704b4cb6ae4ec73522efe05f/file: No such file or directory
```

Pachyderm won't let you read data from a commit until the commit is finished.
This prevents reads from racing with other writes and makes every write to the
filesystem atomic. Let's finish the commit:

```shell
$ pachctl finish-commit foo 6a7ddaf3704b4cb6ae4ec73522efe05f
```

Now we can view the files:

```shell
$ cat /pfs/foo/6a7ddaf3704b4cb6ae4ec73522efe05f/file
foo
```

However we've lost the ability to write to it, finished commits are immutable.

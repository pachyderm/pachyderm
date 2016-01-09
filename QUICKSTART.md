# Quick Start

### Prerequisites

- Docker >= 1.9 (must deploy with [`--storage-driver=devicemapper`](http://muehe.org/posts/switching-docker-from-aufs-to-devicemapper/))
- Go >= 1.5
- Kubernetes and Kubectl >= 1.1.2

### Forward Ports

You'll need to make sure `pachctl` can connect to the running pachyderm services:

```shell
$ ssh KUBEHOST -fTNL 650:localhost:30650 -L 651:localhost:30651
```

#### If you have your own Kubernetes

And `kubectl` works, run:

```shell
$ make launch
```

#### If you don't have your Kubernetes

You can launch a dev cluster which runs kubernetes on docker:

```shell
$ make launch-dev
```

This should create a new cluster, to check if it worked do:

```shell
$ kubectl get svc
NAME         CLUSTER_IP   EXTERNAL_IP   PORT(S)                        SELECTOR      AGE
etcd         10.0.0.210   <none>        2379/TCP,2380/TCP              app=etcd      41s
kubernetes   10.0.0.1     <none>        443/TCP                        <none>        3h
pfsd         10.0.0.148   nodes         650/TCP,750/TCP                app=pfsd      41s
ppsd         10.0.0.69    nodes         651/TCP,751/TCP                app=ppsd      41s
rethink      10.0.0.126   <none>        8080/TCP,28015/TCP,29015/TCP   app=rethink   41s
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
$ pachctl create-repo data
$ pachctl create-repo output
$ ls /pfs
data output
```

### Creating a `Commit`
Now that you've created a `Repo` you should see 2 empty directories `/pfs/data`
and `/pfs/output` if you try writing to it, it will fail because you can't write directly to `Repos`, you have to create a
`Commit` to write to instead:

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

a new directory has been created for our commit. This we can write to:

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

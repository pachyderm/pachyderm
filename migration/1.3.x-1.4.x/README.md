## Intro

Pachyderm 1.4 has a different data model from the 1.3 series of releases
and as such it's not possible to map 1.3 data and pipelines 1-to-1 with
equivalents in 1.4. The biggest change between 1.3 and 1.4 concerns large
files, in 1.3 it was possible to store very large files and parallelize
directly over them. While this was convenient for a number of cases it
wound up iteracting in confusing ways with the filesystem layouts we use
for data processing and with our incrementality features. It also felt
a little too magical because it abstracted key details of how computations
would be parallelized away from the user. In 1.4 files can still be large
but large files can only be processed by a single process. If data needs
to be processed in parallel it must be spit into multiple smaller files,
to help with this syntax has been added to the API to help split files.

## Data

1.4 adds the ability to use another PFS instance as a source for
a `put-file` request. This allows you to migrate data from a running 1.3
cluster into a new 1.4 cluster without bringing it down. The first step is
getting both clusters running, if you're running them in the same
Kubernetes cluster namespaces can be useful for avoiding name conflicts.
Once both clusters are up and running you'll need to find an address for
your 1.3 Pachyderm service so you can give it to the 1.4 cluster. This
should be readily via `kubectl get all --namespace=<1.3 namespace>` Then
to put data from a 1.3 repo into a 1.4 repo do:

```
pachctl put-file <1.4 repo> <1.4 branch> -f <1.3 ip address>:30650/<1.3 repo>/<1.3 branch> -r
```

Notice the `-r` parameter which tells Pachyderm to recursively scan the
contents of the branch rather than just putting a single file, you can
also migrate one file at a time by specifying it after the branch
parameter in the address.

Note that now is the time that you need to think about how you want to
parallelize over your data. If your repo contains large files which were
processed in 1.3 with `BLOCK` or `MAP` pipelines then you should use the
`--split` flag in your `put-file` invocation to split the data up now.
Refer to the [`put-file` docs](PLACEHOLDER) to learn more about `--split` works.

After data has been migrated you can safely turn down the 1.3 cluster.

## Pipelines

### Inputs

By far the biggest change to the Pipeline Spec is to the `inputs` field. Inputs
no longer have a `method` field, instead they specify how their inputs can be
parallelized using a glob pattern. [Read more about glob patterns
here](PLACEHOLDER).

The Partitions available in 1.3 match to glob patterns like so:

| Partition | Glob |
| --------- | ---- |
| `BLOCK`   | /*/* |
| `FILE`    | /*   |
| `REPO`    | /    |

Note that to get the equivalent of `BLOCK` partition you should have put
your files with `--split` in the step above. If your files are in a more
deeply nested directory structure you'll need more layers of `*`s in your
glob pattern. I.e. if you put the file `dir/foo` using `--split` rather
than the file `foo` then you'll want to use `/*/*/*`.

In 1.3 it wasn't possible to do `FILE` partitioning below the root layer
of directories, in 1.4 this is possible with glob patterns. I.e. in the
above example where a file is put to `dir/foo` with `--split` you could
process all the pieces of the file together by passing `/*/*`.

1.4 currently does not have a notion of incrementality, much of the
incremental functionality present in 1.3 currently happens automatically.
We'll be introducing a more advanced form of incrementality in a later
version of 1.4 for incremental workloads like sum operations.

In 1.4 `input`s are targeted at specific branches rather than all branches
in a repo, this field defaults to master if left blank.

### Output

In 1.4 the `output` field has been renamed to `egress` to avoid confusion
with the pfs output repo, the semantics are unchanged.

Pipelines in 1.4 output to a specific branch, rather than uuid branches.
As with `input`s the output branch can be left blank in which case it will
default to `master`.

### GC POLICY

The `gc_policy` field is no longer necessary in 1.4 because the processing
containers for pipelines are left running between jobs. This results in
lower latency for jobs and less load put on kubernetes.

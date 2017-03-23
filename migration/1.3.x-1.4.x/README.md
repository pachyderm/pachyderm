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
`--split` flag in your `put-file` invocation the split the data up now.

## pachctl garbage-collect

Garbage collect unused data.

### Synopsis

Garbage collect unused data.

When a file/commit/repo is deleted, the data is not immediately removed from
the underlying storage system (e.g. S3) for performance and architectural
reasons.  This is similar to how when you delete a file on your computer, the
file is not necessarily wiped from disk immediately.

To actually remove the data, you will need to manually invoke garbage
collection with "pachctl garbage-collect".

Currently "pachctl garbage-collect" can only be started when there are no
pipelines running.  You also need to ensure that there's no ongoing "put file".
Garbage collection puts the cluster into a readonly mode where no new jobs can
be created and no data can be added.

Pachyderm's garbage collection uses bloom filters to index live objects. This
means that some dead objects may erronously not be deleted during garbage
collection. The probability of this happening depends on how many objects you
have; at around 10M objects it starts to become likely with the default values.
To lower Pachyderm's error rate and make garbage-collection more comprehensive,
you can increase the amount of memory used for the bloom filters with the
--memory flag. The default value is 10MB.


```
pachctl garbage-collect [flags]
```

### Options

```
  -h, --help            help for garbage-collect
  -m, --memory string   The amount of memory to use during garbage collection. Default is 10MB. (default "0")
```

### Options inherited from parent commands

```
      --no-color   Turn off colors.
  -v, --verbose    Output verbose logs
```


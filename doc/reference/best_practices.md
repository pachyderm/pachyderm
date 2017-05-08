# Best Practices

This document discusses best practices for common use cases.

## Shuffling files

Certain pipelines simply shuffle files around.  If you find yourself writing a pipeline that does a lot of copying, such as [Time Windowing](http://docs.pachyderm.io/en/latest/cookbook/time_windows.html), it probably falls into this category.

The best way to shuffle files, especially large files, is to create **symlinks** in the output directory that point to files in the input directory.

For instance, to move a file `log.txt` to `logs/log.txt`, you might be tempted to write a [`transform`](http://pachyderm.readthedocs.io/en/latest/reference/pipeline_spec.html#transform-required) like this:

```sh
cp /pfs/input/log.txt /pfs/out/logs/log.txt
```

However, it's more efficient to create a symlink:

```sh
ln -s /pfs/input/log.txt /pfs/out/logs/log.txt
```

Under the hood, Pachyderm is smart enough to recognize that the output file simply symlinks to a file that already exists in Pachyderm, and therefore skips the upload altogether.

Note that if your shuffling pipeline only needs the names of the input files but not their content, you can use [`lazy input`](http://pachyderm.readthedocs.io/en/latest/reference/pipeline_spec.html#atom-input).  That way, your shuffling pipeline can skip both the download and the upload.

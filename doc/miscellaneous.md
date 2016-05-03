# Miscellaneous Docs

Following are various miscellaneous documents that are too small to exist on their own.  These documents tend to be advanced and specific in nature; if you are just getting started with Pachyderm, you probably don't need to read them.

## Flash-crowd behavior

In distributed systems, a flash-crowd behavior occurs when a large number of nodes send traffic to a particular node in an uncoordinated fashion, causing the node to become a hotspot, resulting in performance degradation.

To understand how such a behavior can occur in Pachyderm, it's important to understand the way requests are sharded in a Pachyderm cluster.  Pachyderm currently employs a simple sharding scheme that shards based on file names.  That is, requests pertaining to a certain file will be sent to a specific node.  As a result, if you have a number of nodes processing a large dataset in parallel, it's advantageous for them to process files in a random order.

For instance, imagine that you have a dataset that contains `file_A`, `file_B`, and `file_C`, each of of which is 1TB in size.  Now, each of your nodes will get a portion of each of these files.  If your nodes independently start processing files in alphanumeric order, they will all start with `file_A`, causing all traffic to be sent to the node that handles `file_A`.  In contrast, if your nodes process files in a random order, traffic will be distributed between three nodes.

## File Removal

Imagine three jobs executing the following commands in parallel:

```shell
rm -rf /pfs/out/file
echo foo > /pfs/out/file
```

What's going to end up in `/pfs/out/file`?  If Pachyderm was implemented naively, the content of `/pfs/out/file` would depend on the specific order in which PFS receives the commands, such that the content could be anything between `foo`, `foofoo`, and `foofoofoo`.

As such, Pachyderm implements file removal such that it only affects content written prior to the current commit.  As a result, the `rm` command above will **not** remove content being written in the current commit by `echo`.  Rather, it only removes content that exists before this commit.

Therefore, `/pfs/out/file` will always end up with `foofoofoo` in the scenario above.

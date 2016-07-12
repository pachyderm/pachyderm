Pachyderm Pipeline System (PPS)
===============================

- Primitives
  - job
    - mounting settings / access
      - mounting repo -- incremental vs noy
    - environment variables
  - pipeline
    - images / hosting / private / etc
    - parallelism
    - strategies / partitn w examples!
- Provenance
- Debugging tools
- Scaling

___

Pachyderm Pipeline System is a parallel containerized analysis platform

It is designed to:

- write your analysis in any language of your choosing ([enabling Accessibility](https://pachyderm.io/dsbor.html)).
- allow you to compose your analyses
- allow you to reproduce your input data, your processing step, and your output data ([enabling Reproducibility](https://pachyderm.io/dsbor.html))
- allow you to understand the [Provenance](#provenance) of your data

# Primitives

Pachyderm




###################################################

# Strategy

Map vs Reduce

# Incrementality

Determines how the input repo is mounted. 




### On a PPS Job

On a [PPS Job](#pachyderm_pipeline_system.html#Job) your files are mounted a bit differently.


- where its mounted
- conventions about output/input
- how your data is exposed


## Flash-crowd behavior

In distributed systems, a flash-crowd behavior occurs when a large number of nodes send traffic to a particular node in an uncoordinated fashion, causing the node to become a hotspot, resulting in performance degradation.

To understand how such a behavior can occur in Pachyderm, it's important to understand the way requests are sharded in a Pachyderm cluster.  Pachyderm currently employs a simple sharding scheme that shards based on file names.  That is, requests pertaining to a certain file will be sent to a specific node.  As a result, if you have a number of nodes processing a large dataset in parallel, it's advantageous for them to process files in a random order.

For instance, imagine that you have a dataset that contains `file_A`, `file_B`, and `file_C`, each of of which is 1TB in size.  Now, each of your nodes will get a portion of each of these files.  If your nodes independently start processing files in alphanumeric order, they will all start with `file_A`, causing all traffic to be sent to the node that handles `file_A`.  In contrast, if your nodes process files in a random order, traffic will be distributed between three nodes.


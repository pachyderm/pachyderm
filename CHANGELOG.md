# Changelog

## 1.2.0

Features:

- PFS has been rewritten to be more reliable and optimizeable
- PFS now has a much simpler name scheme for commits (eg `master/10`)
- PFS now supports merging, there are 2 types of merge. Squash and Replay
- Caching has been added to several of the higher cost parts of PFS
- UpdatePipeline, which allows you to modify an existing pipeline
- Transforms now have an Env section for specifying environment variables
- ArchiveCommit, which allows you to make commits not visible in ListCommit but still present and readable
- ArchiveAll, which archives all data
- PutFile can now take a URL in place of a local file, put multiple files and start/finish its own commits
- Incremental Pipelines now allow more control over what data is shown
- `pachctl deploy` is now the recommended way to deploy a cluster
- `pachctl port-forward` should be a much more reliable way to get your local machine talking to pachd
- `pachctl mount` will recover if it loses and regains contact with pachd
- `pachctl unmount` has been added, it can be used to unmount a single mount or all of them with `-a`
- Benchmarks have been added
- pprof support has been added to pachd
- Parallelization can now be set as a factor of cluster size
- `pachctl put-file` has 2 new flags `-c` and `-i` that make it more usable
- Minikube is now the recommended way to deploy locally

Content:

- Our developer portal is now available at: http://pachyderm.readthedocs.io/en/latest/
- We've added a quick way for people to reach us on Slack at: http://slack.pachyderm.io
- OpenCV example

## 1.1.0

Features:

- Data Provenance, which tracks the flow of data as it's analyzed
- FlushCommit, which tracks commits forward downstream results computed from them
- DeleteAll, which restores the cluster to factory settings
- More featureful data partitioning (map, reduce and global methods)
- Explicit incrementality
- Better support for dynamic membership (nodes leaving and entering the cluster)
- Commit IDs are now present as env vars for jobs
- Deletes and reads now work during job execution
- pachctl inspect-* now returns much more information about the inspected objects
- PipelineInfos now contain a count of job outcomes for the pipeline
- Fixes to pachyderm and bazil.org/fuse to support writing a larger number of files
- Jobs now report their end times as well as their start times
- Jobs have a pulling state for when the container is being pulled
- Put-file now accepts a -f flag for easier puts
- Cluster restarts now work, even if kubernetes is restarted as well
- Support for json and binary delimiters in data chunking
- Manifests now reference specific pachyderm container version making deployment more bulletproof
- Readiness checks for pachd which makes deployment more bulletproof
- Kubernetes jobs are now created in the same namespace pachd is deployed in
- Support for pipeline DAGs that aren't transitive reductions.
- Appending to files now works in jobs, from shell scripts you can do `>>`
- Network traffic is reduced with object stores by taking advantage of content addressability
- Transforms now have a `Debug` field which turns on debug logging for the job
- Pachctl can now be installed via Homebrew on macOS or apt on Ubuntu
- ListJob now orders jobs by creation time
- Openshift Origin is now supported as a deployment platform

Content:

- Webscraper example
- Neural net example with Tensor Flow
- Wordcount example

Bug fixes:

- False positive on running pipelines
- Makefile bulletproofing to make sure things are installed when they're needed
- Races within the FUSE driver
- In 1.0 it was possible to get duplicate job ids which, that should be fixed now
- Pipelines could get stuck in the pulling state after being recreated several times
- Map jobs no longer return when sharded unless the files are actually empty
- The fuse driver could encounter a bounds error during execution, no longer
- Pipelines no longer get stuck in restarting state when the cluster is restarted
- Failed jobs were being marked failed too early resulting in a race condition
- Jobs could get stuck in running when they had failed
- Pachd could panic due to membership changes
- Starting a commit with a nonexistant parent now errors instead of silently failing
- Previously pachd nodes would crash when deleting a watched repo
- Jobs now get recreated if you delete and recreate a pipeline
- Getting files from non existant commits gives a nicer error message
- RunPipeline would fail to create a new job if the pipeline had already run
- FUSE no longer chokes if a commit is closed after the mount happened
- GCE/AWS backends have been made a lot more reliable

Tests:

From 1.0.0 to 1.1.0 we've gone from 70 tests to 120, a 71% increase.

## 1.0.0 (5/4/2016)

1.0.0 is the first generally available release of Pachyderm.
It's a complete rewrite of the 0.* series of releases, sharing no code with them.
The following major architectural changes have happened since 0.*:

- All network communication and serialization is done using protocol buffers and GRPC.
- BTRFS has been removed, instead build on object storage, s3 and GCS are currently supported.
- Everything in Pachyderm is now scheduled on Kubernetes, this includes Pachyderm services and user jobs.
- We now have several access methods, you can use `pachctl` from the command line, our go client within your own code and the FUSE filesystem layer

# Changelog

## 1.8.0

- Completely rewritten hashtree backend that provides massive performance boosts.
- Single sign-on Auth via Okta.
- Support for groups and robot users.
- Support for splitting file formats with headers and footers such as SQL and CSV.

## 1.7.10

- Adds `put-file --split` support for SQL dumps. (#3064)
- Adds support for headers and footers for data types passed to `--split` such as CSV and the above mentioned SQL. (#3064)
- Adds support for accessing previous versions of pipelines using the same syntax as is used with commits. I.e. `pachctl inspect-pipeline foo^` will give the previous version of `foo`. (#3159)
- Adds support in pipelines for additional Kubernetes primitives on workers, including: node selectors, priority class and storage requests and limits. Additionally there is now a field in the pipeline spec `pod_spec` that allows you to set any field on the pod using json. (#3169)

## 1.7.9

- Moves garbage collection over to a bloom filter based indexing method. This
greatly decreases the amount of memory that garbage collection requires, at the
cost of a small probability of not deleting objects that should be. Garbage
collection can be made more accurate by using more memory with the flag
`--memory` passed to `pachctl garbage-collect`. (#3161)

## 1.7.8

- Fixes multiple issues that could cause jobs to hang when they encountered intermittent errors such as network hiccups. (#3155)

## 1.7.7

- Greatly improves the performance of the pfs FUSE implementation. Performance should be close to on par with the that of pachctl get-file. The only trade-off is that the new implementation will use disk space to cache file contents. (#3140)

## 1.7.6

- Pachyderm's FUSE support (`pachctl mount`) has been rewritten. (#3088)
- `put-file` requests that put files from multiple sources (`-i` or `-r`) now create a single commit. (#3118) 
- Fixes a bug that caused `put-file` to throw spurious warnings about URL formatted paths. (#3117)
- Several fixes have been made to which user code runs as to allow services such as Jupyter to work out of the box. (#3085)
- `pachctl` now has `auth set-config` and `auth get-config` commands. (#3095)<Paste>

## 1.7.5

- Workers no longer run privileged containers. (#3031) To achieve this a few modifications had to be made to the `/pfs` directory that may impact some user code. Directories under `/pfs` are now symlinks to directories, previously they were bind-mounts (which requires that the container be privileged). Furthermore there's now a hidden directory under `/pfs` called `.scratch` which contains the directories that the symlinks under `/pfs` point to.
- The number of times datums are retries is now configurable. (#3033)
- Fixed a bug that could cause Kubernetes errors to prevent pipelines from coming up permanently. (#3043, #3005)
- Robot users can now modify admins. (#3049)
- Fixed a bug that could permanently lock robot-only admins out of the cluster. (#3050)
- Fixed a couple of bugs (#3045, #3046) that occurred when a pipeline was rapidly updated several times. (#3054)
- `restore` now propagates user credentials, allowing it to work on clusters with auth turned on. (#3057)
- Adds a `debug-dump` command which dumps running goroutines from the cluster. (#3078)
- `pachd` now prints a full goroutine dump if it encounters an error. (#3103)

## 1.7.4

- Fixes a bug that prevented image pull secrets from propagating through `pachctl deploy`. (#2956, thanks to @jkinkead)
- Fixes a bug that made `get-file` fail on empty files. (#2960)
- `ListFile` and `GlobFile` now return results leixcographically sorted. (#2972)
- Fixes a bug that caused `Extract` to crash. (#2973)
- Fixes a bug that caused pachd to crash when given a pipeline without a name field. (#2974)
- Adds dial options to the Go client's connect methods. (#2978)
- `pachctl get-logs` now accepts `-p` as a synonym for `--pipeline`. (#3009, special thanks to @jdelfino)
- Fixes a bug that caused connections to leak in the vault plugin. (#3016)
- Fixes a bug that caused incremental pipelines that are downstream from other pipelines to not run incrementally. (#3023)
- Updates monitoring deployments to use the latest versions of Influx, Prometheus and Grafana. (#3026)
- Fixes a bug that caused `update-pipeline` to modify jobs that had already run. (#3028)

## 1.7.3

- Fixes an issue that caused etcd deployment to fail when using a StatefulSet. (#2929, #2937)
- Fixes an issue that prevented pipelines from starting up. (#2949)

## 1.7.2

- Pachyderm now exposes metrics via Prometheus. (#2856)
- File commands now all support globbing syntax. I.e. you can do pachctl list-file ... foo/*. (#2870)
- garbage-collect is now safer and less error prone. (#2912)
- put-file no longer requires starting (or finishing) a commit. Similar to put-file -c, but serverside. (#2890)
- pachctl deploy --dry-run can now output YAML as well as JSON. Special thanks to @jkinkead. (#2872)
- Requirements on pipeline container images have been removed. (#2897)
- Pachyderm no longer requires privileged pods. (#2887)
- Fixes several issues that prevented deleting objects in degraded states. (#2912)
- Fixes bugs that could cause stats branches to not be cleaned up. (#2855)
- Fixes 2 bugs related to auth services not coming up completely. (#2843)
- Fixes a bug that prevented pachctl deploy storage amazon from working. (#2863)
- Fixes a class of bugs that occurred due to misuse of our collections package. (#2865)
- Fixes a bug that caused list-job to delete old jobs if you weren't logged in. (#2879)
- Fixes a bug that caused put-file --split to create too many goroutines. (#2906)
- Fixes a bug that prevent deploying to AWS using an IAM role. (#2913)
- Pachyderm now deploys and uses the latest version of etcd. (#2914)

## 1.7.1

- Introduces a new model for scaling up and down pipeline workers. [Read more](http://docs.pachyderm.io/en/latest/reference/pipeline_spec.html#standby-optional).
- It's now possible to run Pachyderm without workers needing access to the docker socket. (#2813)
- Fixes a bug that caused stats enabled pipelines to get stuck in a restart loop if they were deleted and recreated. (#2816)
- Fixes a bug that broke logging due to removing newlines between log messages. (#2852)
- Fixes a bug that caused pachd to segfault when etcd didn't come up properly. (#2840)
- Fixes a bug that would cause jobs to occasionally fail with a "broken pipe" error. (#2832)
- `pachctl version` now supports the `--raw` flag like other `pachctl` commands. (#2817)
- Fixes a bug that caused `max_queue_size` to be ignored in pipelines. (#2818)

## 1.7.0

- Implements a new algorithm for triggering jobs in response to new commits.
- Pachyderm now tracks subvenance, the inverse of provenance.
- Branches now track provenance and subvenance.
- Restrictions on delete-commit have been removed, you can now delete any input commit and the DAG will repair itself appropriately.
- Pachyderm workers no longer use long running grpc requests to schedule work, they use an etcd based queue instead. This solves a number of bugs we had with larger jobs.
- You can now backup and restore your cluster with extract and restore.
- Pipelines now support timeouts, both for the job as a whole or for individual datums.
- You can now follow jobs logs with -f.
- Support for Kubernetes RBAC.
- Docker images with entrypoints can now be run, you do this by not specifying a cmd.
- Pachctl now has bash completion, including for values stored within it. (pachctl completion to install it)
- pachctl deploy now has a --namespace flag to deploy to a specific namespace.
- You can no longer commit directly to output repos, this would cause a number of problems with the internal model that were tough to recover from.

## 1.6.10

- Fixes a bug in extract that prevented some migrations from completing.

## 1.6.9

- Adds admin commands extract and restore.

## 1.6.8

- Fixed an issue that could cause output data to get doubled. (#2644)
- Fix / add filtering of jobs in list-job by input commits. (#2642)
- Extends bash completion to cover values as well as keywords. (#2617)
- Adds better validation of file paths. (#2627)

## 1.6.7

- Support for Google Service Accounts
- RBAC support
- Follow and tail logs
- Expose public IP for githook service
- Handle many 100k+ files in a single commit, which allows users to more easily manage/version millions of files.
- Fix datum status in the UI

## 1.6.6

- Users can now specify k8s resource limits on a pipeline
- Users can specify a `datum_timeout` and `job_timeout` on a pipeline
- Minio S3V2 support
- New worker model (to eliminate long running grpc calls)

## 1.6.5

- Adds support for Kubernetes 1.8
- Fixes a bug that caused jobs with small numbers of datums not to use available nodes for processing. #2480.

## 1.6.4

## 1.6.3

- Fixes a bug that corrupted large files ingressed from object stores. #2405
- Fixes a migration bug that could get pipelines stuck in a crash loop
- Fixes an issue with pipelines processing old data #2469
- Fixes a bug that allowed paused pipelines to restart themselves.

## 1.6.2

- Changes default memory settings so that Pachyderm works on Minikube out of the box.
- Implements streaming versions of `ListFile` and `GlobFile` which prevents crashing on larger datasets.
- Fixes a race condition with `GetLogs`

## 1.6.1

- Adds support for private registries. (#2360)
- Fixes a bug that prevent cloud front deployments from working. (#2381)
- Fixes a failure that code arise while watching k8s resources. (#2382)
- Uses k8s' Guaranteed QoS for etcd and pachd. (#2368)

## 1.6.0

New Features:

- Cron Inputs
- Access Control Model
- Advanced Statistic tracking for jobs
- Extended UI

## 1.5.3

Bug Fixes:

- Fix an issue that prevented deployment on GCE #2139
- Fix an issue that could cause jobs to hang due to lockups with bind mounts. #2178
- FromCommit in pipelines is now exclusive and able to be used with branch names as well as commit ids. #2180
- Egress was broken for certain object stores, this should be fixed now. #2156

New Features:

- Union inputs can now be given the same name, making union much more ergonomic. #2174
- PutFile now has an `--overwrite` flag which overwrites the previous version of the file rather than appending. #2142
- We've introduce a new type of input, `Cron`, which can be used to trigger pipelines based on time. #2150.

## 1.5.1 / 1.5.2

### Bug Fixes

* A pipeline can get stuck after repeated worker failures.  (#2064)
* `pachctl port-forward` can leave a orphaned process after it exits.  (#2098)
* `alpine`-based pipelines fail to load input data.  (#2118)
* Logs are written to the object store even when stats is not enabled, slowing down the pipeline unnecessarily.  (#2119)

### Features / Improvements

* Pipelines now support the “stats” feature.  See the [docs](http://pachyderm.readthedocs.io/en/latest/reference/pipeline_spec.html#enable-stats-optional) for details.  (#1998)
* Pipeline cache size is now configurable.  See the [docs](http://pachyderm.readthedocs.io/en/latest/reference/pipeline_spec.html#cache-size-optional) for details.  (#2033)
* `pachctl update-pipeline` now **only** process new input data with the new code; the old input data is not re-processed.  If it’s desired that all data are re-processed, use the `--reprocess` flag.  See the [docs](http://pachyderm.readthedocs.io/en/latest/fundamentals/updating_pipelines.html) for details.  (#2034)
* Pipeline workers now support “pipelining”, meaning that they start downloading the next datums while processing the current datum, thereby improving overall throughput.  (#2057)
* The `scaleDownThreshold` feature has been improved such that when a pipeline is scaled down, the remaining worker only takes up minimal system resources.  (#2091)

## 1.5.0

### Bug Fixes

* Downstream repos' provenance is not updated properly when `update-pipeline` changes the inputs for a pipeline. (#1958)
* `pachctl version` blocks when pachctl doesn't have Internet connectivity. (#1971)
* `incremental` misbehaves when files are deeply nested. (#1974)
* An `incremental` pipeline blocks if there's provenance among its inputs. (#2002)
* PPS fails to create subsequent pipelines if any pipeline failed to be created. (#2004)
* Pipelines sometimes reprocess datums that have already been processed. (#2008)
* Putting files into open commits fails silently. (#2014)
* Pipelines with inputs that use different branch names fail to create jobs. (#2015)
* `get-logs` returns incomplete logs.  (#2019)

### Features

* You can now use `get-file` and `list-file` on open commits. (#1943)

## 1.4.8

### Bug Fixes

- Fixes bugs that caused us to swamp etcd with traffic.
- Fixes a bug that could cause corruption to in pipeline output.

### Features

- Readds incremental processing mode
- Adds `DiffFile` which is similar in function to `git diff`
- Adds the ability to use cloudfront as a caching layer for additional scalability on aws.
- `DeletePipeline` now allows you to delete the output repos as well.
- `DeletePipeline` and `DeleteRepo` now support a `--all` flag

### Removed Features

- Removes one-off jobs, they were a rarely used feature and the same behavior can be replicated with pipelines

## 1.4.7

### Bug fixes

* [Copy elision](http://pachyderm.readthedocs.io/en/latest/reference/best_practices.html#shuffling-files) does not work for directories. (#1803)
* Deleting a file in a closed commit fails silently. (#1804)
* Pachyderm has trouble processing large files. (#1819)
* etcd uses an unexpectedly large amount of space. (#1824)
* `pachctl mount` prints lots of benevolent FUSE errors. (#1840)

### New features

* `create-repo` and `create-pipeline` now accept the `--description` flag, which creates the repo/pipeline with a "description" field.  You can then see the description via `inspect-repo/inspect-pipeline`. (#1805)
* Pachyderm now supports garbage collection, i.e. removing data that's no longer referenced anywhere.  See the [docs](http://pachyderm.readthedocs.io/en/latest/reference/best_practices.html#garbage-collection) for details. (#1826)
* Pachyderm now has GPU support!  See the [docs](http://pachyderm.readthedocs.io/en/latest/cookbook/tensorflow_gpu.html) for details. (#1835)
* Most commands in `pachctl` now support the `--raw` flag, which prints the raw JSON data as opposed to pretty-printing.  For instance, `pachctl inspect-pipeline --raw` would print something akin to a pipeline spec. (#1839)
* `pachctl` now supports `delete-commit`, which allows for deleting a commit that's not been finished.  This is useful when you have added the wrong data in a commit and you want to start over.
* The web UI has added a file viewer, which allows for viewing PFS file content in the browser.

## 1.4.6

### Bug fixes

* `get-logs` returns errors along the lines of `Invalid character…`. (#1741)
* etcd is not properly namespaced. (#1751)
* A job might get stuck if it uses `cp -r` with lazy files. (#1757)
* Pachyderm can use a huge amount of memory, especially when it processes a large number of files. (#1762)
* etcd returns `database space exceeded` errors after the cluster has been running for a while. (#1771)
* Jobs crashing might eventually lead to disk space being exhausted. (#1772)
* `port-forward` uses wrong port for UI websocket requests to remote clusters (#1754)
* Pipelines can end up with no running workers when the cluster is under heavy load. (#1788)
* API calls can start returning `context deadline exceeded` when the cluster is under heavy load. (#1796)

### New features / improvements

* Union input: a pipeline can now take the union of inputs, in addition to the cross-product of them.  Note that the old `inputs` field in the pipeline spec has been deprecated in favor of the new `input` field.  See the [pipeline spec](http://pachyderm.readthedocs.io/en/latest/reference/pipeline_spec.html#input-required) for details. (#1665)
* Copy elision: a pipeline that shuffles files can now be made more efficient by simply outputting symlinks to input files.  See the [docs on shuffling files](http://pachyderm.readthedocs.io/en/latest/reference/best_practices.html#shuffling-files) for details. (#1791)
* `pachctl glob-file`: ever wonder if your glob pattern actually works?  Wonder no more.  You can now use `pachctl glob-file` to see the files that match a given glob pattern. (#1795)
* Workers no longer send/receive data through pachd.  As a result, pachd is a lot more responsive and stable even when there are many ongoing jobs.  (#1742)

## 1.4.5

### Bug fixes

* Fix a bug where pachd may crash after creating/updating a pipeline that has many input commits. (#1678)
* Rules for determining when input data is re-processed are made more intuitive.  Before, if you update a pipeline without updating the `transform`, the input data is not re-processed.  Now, different pipelines or different versions of pipelines always re-process data, even if they have the same `transform`. (#1685)
* Fix several issues with jobs getting stuck. (#1717)
* Fix several issues with lazy pipelines getting stuck. (#1721)
* Fix an issue with Minio deployment that results in job crash loop. (#1723)
* Fix an issue where a job can crash if it outputs a large number of files. (#1724)
* Fix an issue that causes intermittent gRPC errors. (#1727)

### New features

* Pachyderm now ships with a web UI!  To deploy a new Pachyderm cluster with the UI, use `pachctl deploy <arguments> --dashboard`.  To deploy the UI onto an existing cluster, use `pachctl deploy <arguments> --dashboard-only`.  To access the UI, simply `pachctl port-forward`, then go to `localhost:38080`.  Note that the web UI is currently in alpha; expect bugs and significant changes.   
* You can now specify the amount of resources (i.e. CPU & memory) used by Pachyderm and etcd.  See `pachctl deploy --help` for details. (#1676)
* You can now specify the amount of resources (i.e. CPU & memory) used by your pipelines.  See the [pipeline spec](http://pachyderm.readthedocs.io/en/latest/reference/pipeline_spec.html#resource-spec-optional) for details. (#1683)

## 1.4.4

### Bug fixes

* A job can fail to restart when encountering an internal error.
* A deployment with multiple pachd nodes can get stalled jobs.
* `delete-pipeline` is supposed to have the `--delete-jobs` flag but doesn't.
* `delete-pipeline` can fail if there are many jobs in the pipeline.
* `update-pipeline` can fail if the original pipeline has not outputted any commits.
* pachd can crash if etcd is flaky.
* pachd memory can be easily exhausted on GCE deployments.
* If a pipeline is created with multiple input commits already present, all jobs spawn and run in parallel.  After the fix, jobs always run serially.

### Features

* Pachyderm now supports auto-scaling: a pipeline's worker pods can be terminated automatically when the pipeline has been idle for a configurable amount of time.  See the `scaleDownThreshold` field of the [pipeline spec](http://pachyderm.readthedocs.io/en/latest/reference/pipeline_spec.html#scale-down-threshold-optional) for details.
* The processing of a datum can be restarted manually via `restart-datum`.
* Workers' statuses are now exposed through `inspect-job`.
* A job can be stopped manually via `stop-job`.

## 1.4.3

### Bug fixes

* Pipelines with multiple inputs process only a subset of data.
* Workers may fall into a crash loop under certain circumstances. (#1606)

### New features

* `list-job` and `inspect-job` now display a job's progress, i.e. they display the number of datums processed thus far, and the total number of datums.
* `delete-pipeline` now accepts an option (`--delete-jobs`) that deletes all jobs in the pipeline. (#1540)
* Azure deployments now support dynamic provisioning of volumes.

## 1.4.2

### Bug fixes

* Certain network failures may cause a job to be stuck in the `running` state forever.
* A job might get triggered even if one of its inputs is empty.
* Listing or getting files from an empty output commit results in `node "" not found` error.
* Jobs are not labeled as `failure` even when the user code has failed.
* Running jobs do not resume when pachd restarts.
* `put-file --recursive` can fail when there are a large number of files.
* minio-based deployments are broken.

### Features

* `pachctl list-job` and `pachctl inspect-job` now display the number of times each job has restarted.
* `pachctl list-job` now displays the pipeline of a job even if the job hasn't completed.

## 1.4.1

### Bug fixes

* Getting files from GCE results in errors.
* A pipeline that has multiple inputs might place data into the wrong `/pfs` directories.
* `pachctl put-file --split` errors when splitting to a large number of files.
* Pipeline names do not allow underscores.
* `egress` does not work with a pipeline that outputs a large number of files.
* Deleting nonexistent files returns errors.
* A job might try to process datums even if the job has been terminated.
* A job doesn't exit after it has encountered a failure.
* Azure backend returns an error if it writes to an object that already exists.

### New features

* `pachctl get-file` now supports the `--recursive` flag, which can be used to download directories.
* `pachctl get-logs` now outputs unstructured logs by default.  To see structured/annotated logs, use the `--raw` flag.

## 1.4.0

Features/improvements:

- Correct processing of modifications and deletions.  In prior versions, Pachyderm pipelines can only process data additions; data that are removed or modified are effectively ignored.  In 1.4, when certain input data are removed (or modified), downstream pipelines know to remove (or modify) the output that were produced as a result of processing the said input data.

As a consequence of this change, a user can now fix a pipeline that has processed erroneous data by simply making a new commit that fixes the said erroneous data, as opposed to having to create a new pipeline.

- Vastly improved performance for metadata operations (e.g. list-file, inspect-file).  In prior versions, metadata operations on commits that are N levels deep are O(N) in runtime.  In 1.4, metadata operations are always O(1), regardless of the depth of the commit.

- A new way to specify how input data is partitioned.  Instead of using two flags `partition` and `incrementality`, we now use a single `glob` pattern.  See the [glob doc](http://pachyderm.readthedocs.io/en/stable/reference/pipeline_spec.html#input-glob-pattern) for details.

- Flexible branch management.  In prior versions, branches are fixed, in that a commit always stays on the same branch, and a branch always refers to the same series of commits.  In 1.4, branches are modeled similar to Git's tags; they can be created, deleted, and renamed independently of commits.

- Simplified commit states.  In prior versions, commits can be in many states including `started`, `finished`, `cancelled`, and `archived`.  In particular, `cancelled` and `archived` have confusing semantics that routinely trip up users.  In 1.4, `cancelled` and `archived` have been removed.

- Flexible pipeline updates.  In prior versions, pipeline updates are all-or-nothing.  That is, an updated pipeline either processes all commits from scratch, or it processes only new commits.  In 1.4, it's possible to have the updated pipeline start processing from any given commit.

- Reduced cluster resource consumption.  In prior versions, each Pachyderm job spawns up a Kubernetes job which in turn spawns up N pods, where N is the user-specified parallelism.  In 1.4, all jobs from a pipeline share N pods.  As a result, a cluster running 1.4 will likely spawn up way fewer pods and use fewer resources in total.

- Simplified deployment dependencies.  In prior versions, Pachyderm depends on RethinkDB and etcd to function.  In 1.4, Pachyderm no longer depends on RethinkDB.

- Dynamic volume provisioning.  GCE and AWS users (Azure support is coming soon) no longer have to manually provision persistent volumes for deploying Pachyderm.  `pachctl deploy` is now able to dynamically provision persistent volumes.  See the [deployment doc](http://pachyderm.readthedocs.io/en/stable/deployment/deploy_intro.html) for details.

Removed features:

A handful of APIs have been removed because they no longer make sense in 1.4.  They include:

- ForkCommit (no longer necessary given the new branch APIs)
- ArchiveCommit (the `archived` commit state has been removed)
- ArchiveAll (same as above)
- DeleteCommit (the original implementation of DeleteCommit is very limiting: only open head commits may be removed.  An improved version of DeleteCommit is coming soon)
- SquashCommit (was only necessary due to the way PPS worked in prior versions)
- ReplayCommit (same as above)

## 1.3.0

Features:

- Embedded Applications - Our “service” enhancement allows you to embed applications, like Jupyter, dashboards, etc., within Pachyderm, access versioned data from within the applications, and expose the applications externally.
- Pre-Fetched Input Data - End-to-end performance of typical Pachyderm pipelines will see a many-fold speed up thanks to a prefetch of input data.
- Put Files via Object Store URLs - You can now use “put-file” with s3://, gcs://, and as:// URLS.
- Update your Pipeline code easily - You can now call “create-pipeline” or “update-pipeline” with the “--push-images” flag to re-run your pipeline on the same data with new images.
- Support for all Docker images - It is no longer necessary to include anything Pachyderm specific in your custom Docker images, so use any Docker image you like (with a couple very small caveats discussed below).
- Cloud Deployment with a single command for Amazon / Google / Microsoft / a local cluster - via `pachctl deploy ...`
- Migration support for all Pachyderm data from version `1.2.2` through latest `1.3.0`
- High Availability upgrade to rethink, which is now deployed as a petset
- Upgraded fault tolerance via a new PPS job subscription model
- Removed redundancy in log messages, making logs substantially smaller
- Garbage collect completed jobs
- Support for deleting a commit
- Added user metrics (and an opt out mechanism) to anonymously track usage, so we can discover new bottlenecks
- Upgrade to k8s 1.4.6

## 1.2.0

Features:

- PFS has been rewritten to be more reliable and optimizeable
- PFS now has a much simpler name scheme for commits (e.g. `master/10`)
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
- Starting a commit with a nonexistent parent now errors instead of silently failing
- Previously pachd nodes would crash when deleting a watched repo
- Jobs now get recreated if you delete and recreate a pipeline
- Getting files from non existent commits gives a nicer error message
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

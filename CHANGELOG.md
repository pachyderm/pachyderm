# Changelog

## 1.11.7
- Changes to fix Jaeger tracing functionality (#5331)
- Reverted a change that accidentally made storage credentials required in custom deployment when upgrading to 1.11.6 (#5421)

## 1.11.6
- Added a deploy option to enable verbose logging in S3 client (#5340)
- Fix a bug that would leak a revoked pipeline token object (#5397)
- Fix a bug causing extra data to be written to small job artifact files in some cases (#5401)
- Fix a bug causing workers to attempt to read certain job artifacts before they were fully written (#5402)
- Added support to display when a job is in the egress state (#5411)

## 1.11.5
- Changes to fix multiple error log messages when processing `list pipeline` (#5304)
- Fixes a bug that can cause get file request to fail when the request falls on a certain boundary condition (#5316)
- Fixes a bug that can leave a stats commit open when stats enabled pipeline is updated with `--reprocess` option. This bug will also prevent new jobs from getting created. (#5321)
- Changes for better error handling when pipelines info cannot be fully initialized due to transient failures or user errors (#5322)
- Fixes a bug that did not stop a job before deleting a job when `delete job` is called (#5326)
- Fixes a family of bugs to handle pipeline state transitions. The change resolves a few issues: pipelines getting stuck in STARTING state if Kubernetes is unavailable; cannot delete and recreate pipelines in STANDBY state; fixes jobs occasionally getting stuck in CRASHING state (#5330) (#5357)
- Fixes a family of bug that did not properly clean up temporary artifacts from a job (#5332)
- Changes to move some noisy log message to DEBUG level (#5352)
- Fixes a bug that can sometimes leave pipeline in STANDBY state (#5364)
- Fixes a bug that causes incorrect datums to be processed due to trailing slashes in joins (#5366)
- Changes the metric reporting interval to 60mins (#5375)
- Fixes a bug that loses auth credentials from pipelines after 30 days (#5388)

## 1.11.4
- Fixes a race condition that prevents a standby pipeline from transitioning out of crashing state (#5273)
- Fixes a bug that leaked goroutine (#5288)
- Fixes a bug that did not correctly set the provenance when specified in `run pipeline` command (#5299)

## 1.11.3
- Fixes a bug that did not correctly port forward OIDC port (#5221)
- Changes to allow configuration of SAML and OIDC default server ports (#5234)
- Changes to improve the reliability of handling streams in spouts (#5240)
- Fixes a bug that would fail the `run cron <pipeline>` command if multiple cron inputs have been specified (#5241)

## 1.11.2
- Changes to create/update pipeline to warn users about using the “latest” tag in their images (#5164)
- Fixes a bug that mistagged user logs messages for spouts and services as master log messages (#5187)
- Fixed a bug that would return an error when listing commits and the list reaches the user-specified limit (#5190)
- Fixes `create_python_pipeline` in the python client library when auth is enabled (#5194)
- Fixes a bug that fails to execute a pipeline if the build pipeline does not any wheels (#5197)
- Fixes a bug that would immediately cancel job egress (#5201)
- Fixes a bug that prevented progress counts from being updated. In addition, make progress counts update more granularly in `inspect job` (#5206)
- Fixes a bug that would cause certain kinds of jobs to pick an incorrect commit if there were multiple commits on the same branch in the provenance (#5207)

## 1.11.1
- Fixed a race condition that updated a job state after it is finished (#5099)
- Fixes a bug that would prevent successful initialization (#5130)
- Changes to `debug dump` command to capture debug info from pachd and all worker pods by default. Debug info includes logs, goroutines, profiles, and specs (#5150)

## 1.11.0

Deprecation notice: Support for S3V2 signatures is deprecated in 1.11.0 and will reach end-of-life in 1.12.0. Users who are using S3V4-capable storage should make sure their deployment is using the supported storage backend by redeploying without `--isS3V2` flag. If you need help, please reach out to Pachyderm support.

- Adds support for running multiple jobs in parallel in a single pipeline (#4572)
- Adds support for logs stack traces when a request encounters an error (#4681)
- Adds support for the first release of the pachyderm IDE (#4732) (#4790) (#4838) 
- Adds support for displaying progress bar during `pachctl put file` (#4745)
- Adds support for writable pachctl mount which checkpoints data back into pfs when it's unmounted (#4772)
- Adds support for metric endpoint configurable via METRICS_ENDPOINT env variable (#4793) 
- Adds an "exit" command to the pachctl shell (#4802)
- Adds a `--compress` option to `pachctl put file` which GZIP compresses the upload stream (#4814)
- Adds a `--put-file-concurrency-limit` option to `pachctl put file` command to limits the upload parallelism which limits the memory footprint in pachd to avoid OOM condition (#4827) 
- Adds support to periodically reload TLS certs (#4835)
- Adds a new pipeline state "crashing" which pipelines enter when they encounter Kubernetes errors. Pipelines in this state will have a human-readable "Reason" that explains why they're crashing. Pipelines also now expose the number of pods that are up and responsive. Both values can be seen with `inspect pipeline` (#4922)
- Adds support to allow etcd volumes to be expanded. (Special thanks to @mattrobenolt.) (#4925)
- Adds experimental support for using Loki as a logging backend rather than k8s. Enable with the `LOKI_LOGGING` feature flag to pachd (#4946)
- Adds support for copy object in S3 gateway (#4972)
- Adds a new cluster-admin role, "FS", which grants access to all repos but not other admin-only endpoints (#4975) (#5103)
- Adds support to surface image pull errors in pipeline sidecar containers (#4979)
- Adds support for colorizing level in `pachctl logs` (Special thanks to @farhaanbukhsh) (#4996)
- Adds configurable resource limits to the storage side and set default resource limits for the init container (#4999) 
- Adds support user sign in by authenticating with an OIDC provider (#5005)
- Adds error handling when starting a transaction when another one is pending (Special thanks to @farhaanbukhsh) (#5010)
- Adds support for using TLS (if enabled) for downloading files over HTTP (#5023)
- Adds an option for specifying the Kubernetes service account to use in worker pods (#5056)
- Adds build steps for pipelines (#5064)
- Adds support for a dockerized version of `pachctl` available on docker hub (#5073) (#5079)
- Adds support for configuring Go's GC Percentage (#5089)
- Changes to propagate feature flags to sidecar (#4718) 
- Changes to route all object store access through the sidecar (#4741) 
- Changes to better support disparate S3 client behaviors. Includes numerous compatibility improvements in S3 gateway (#4902) 
- Changes debug dump to collect sidecar goroutines (#4954) 
- Fixes a bug that would cause spouts to lose data when spouts are rapidly opened and closed (#4693) (#4910)
- Fixes a bug that allowed spouts with inputs (#4747) 
- Fixes a bug that prevented access to S3 gateway when other workers are running in a different namespace than Pachyderm namespace (#4753) 
- Fixes a bug that would not delete Kubernetes service when a pipeline is restarted due to updates (#4782) 
- Fixes a bug that created messages larger than expected size which can fail some operations with grpc: received message larger than max error (#4819) 
- Fixes a bug that caused an EOF error in get file request when using azure blob storage client (#4824) 
- Fixes a bug that would fail a restore operation in certain scenarios when the extract operation captures commits in certain failed/incomplete states (#4839) 
- Fixes a bug that causes garbage collection to fail for standby pipelines (#4860) 
- Fixes a bug that did not use the native DNS resolver in pachctl client which may prevent pachd access over VPNs (#4876)  
- Fixes a bug that caused `pachctl list datum <running job>` to return an error "output commit not finished" on pipelines with stats enabled (#4886) 
- Fixes a bug causing a resource leak in pachd when certain protocol errors occur in PutFile (#4908)
- Fixes a bug where downloading files over HTTP didn't work with authorization enabled (#4930)
- Fixes a family of issues that caused workers to indefinitely wait on etcd after a pod eviction (#4947) (#4948) (#4959) 
- Fixes a bug that did not set environment variables for service pipelines (#5009)
- Fixes a bug where users get an error if they run `pachctl debug pprof`, but don't have to “go” installed on their machine (#5022)
- Fixes a bug which caused the metadata of a spout pipeline's spec commit to grow without bound (#5050)
- Fixes a bug that caused the metadata in commit info to not get carried between an extract and a restore operation (#5052)
- Fixes a bug which caused crashes when creating pipelines with certain invalid parameters (#5054)
- Fixes a bug that causes the dash compatibility file not found error (#5063)
- Moves etcd image to Docker Hub from Quay.io (#4899)
- Updates dash version to the latest published version 0.5.48 (#4756)

## 1.10.0

- Change Pachyderm license from Apache 2.0 to Pachyderm Community License
- Changes to how resources are applied to pipeline containers (#4675)
- Changes to GitHook and Prometheus ports (#4537)
- Changes to handle S3 credentials passed to S3 gateway when Auth is disabled (#4585)
- Changes to add support for ‘zsh’ shell (#4494)
- Changes to allow only critical servers to startup with `--required-critical-servers-only` (#4536)
- Changes to improve job logging (#4538)
- Changes to support copying files from output repo to input repos (#4475)
- Changes to ‘flush job’ CLI to support streaming output with --raw option (#4569)
- Changes to remove cluster ID check (#4532)
- Adds annotations and labels to top-level pipeline spec (#4608) (NOTE: If your pipeline spec specifies “service.annotations”, it is recommended that you follow the upgrade path and manually update the pipelines specs to include annotations under the new metadata tag)
- Adds support for S3 inputs & outputs in pipeline specs (#4605, #4660)
- New interactive Pachyderm Shell. The shell provides an easier way to interact with pachctl, including advanced auto-completion support (#4485, #4557)
- Adds support for creating secrets through Pachyderm. (#4483)
- Adds support for disabling commit progress indicator to reduce load on etcd (#4696)
- Fixes a bug that ignored the EDITOR environment variable (#4672)
- Fixes a bug that would cause restore failures from v1.8.x version to v1.9.x+ version (#4662)
- Fixes a bug that would result in missing output data, under specific conditions, when a job resumes processing (#4656)
- Fixes a bug that caused errors when specifying a branch name as the provenance of a new commit (#4657)
- Fixes a bug that would leave a stats commit open under some failure conditions during run pipeline (#4637)
- Fixes a bug that resulted in a stuck merge process when some commits are left in an unfinished state (#4595)
- Fixes a bug that ignored the cron pipeline overwrite value when ‘run cron’ is called from the command line (#4517)
- Fixes a bug that caused `edit pipeline` command to open an empty file (#4526)
- Fixes a bug where some unfinished commit finish times displayed the Unix Epoch time. (#4539)
- Fixes a family of bugs and edge conditions with spout marker (#4487)
- Fixes a bug that would cause crash in ‘diff file’ command (#4601)
- Fixes a bug that caused a crash when `run pipeline` is executed with stats enabled (#4615)
- Fixes a bug that incorrectly skips duplicate datums in a union, under specific conditions (#4691)
- Fixes a bug that ignored the logging level set in the environment variable (#4706) 


## 1.9.12

- New configuration for deployments (exposed through pachctl deploy flags):
  - Only require critical servers to startup and run without error (--require-critical-servers-only). (#4512)
- Improved job logging. (#4523)
- Fixes a bug where some unfinished commit finish times displayed the Unix Epoch time. (#4524) 
- Fixes a bug with edit pipeline. (#4530) 
- Removed cluster id check. (#4534)
- Fixes a bug with spout markers. (#4487)

## 1.9.11

- New configuration for deployments (exposed through pachctl deploy flags):
  - Object storage upload concurrency limit (--upload-concurrency-limit). (#4393)
- Various configuration improvements. (#4442)
- Fixes a bug that would cause workers to segfault. (#4459)
- Upgrades pachyderm to go 1.13.5. (#4472)
- New configuration for amazon and custom deployments (exposed through pachctl deploy amazon/custom flags):
  - Disabling ssl (--disable-ssl) (#4473)
  - Skipping certificate verification (--no-verify-ssl) (#4473)
- Further improves the logging and error reporting during pachd startup. (#4486)
- Removes pprof http server from pachd (debugging should happen through the debug api). (#4496)
- Removes k8s api access from worker code. (#4498)

## 1.9.10

- Fixes a bug that causes `pachctl` to connect to the wrong cluster (#4416)
- Fixes a bug that causes hashtree resource leak in certain conditions (#4420)
- Fixes a family of minor bugs found through static code analysis (#4410)
- Fixes a family of bugs that caused pachd panic when it processed invalid arguments (#4391)
- Fixes a family of bugs that caused deploy yaml to fail (#4290)
- Changes to use standard go modules instead of old vendor directory (#4323)
- Changes to add additional logging during pachd startup (#4447)
- Changes to CLI to add a command, `run cron <pipeline>` to manually trigger a CRON pipeline (#4419)
- Changes to improve performance of join datum processing (#4441)
- Open source Pachyderm S3 gateway to allow applications to interact with PFS storage (#4399)

## 1.9.9

- Adds support for spout marker to keep track of metadata during spout processing. (#4224)
- Updates GPT 2 example to use GPU. (#4325)
- Fixes a bug that did not extract all the pipeline fields (#4204)
- Fixes a bug that did not retry a previously skipped datum when pipeline specs are updated. (#4310)
- Fixes a family of bugs which failed the building of docker images with create pipeline --build command. (#4319)
- Fixed a bug that did not prompt users if auto-derivation of docker credentials fails. (#4319)
- Changes to track commit progress through DAG. (#4203)
- Changes to CLI syntax for run pipeline to accept —job option to re-run a job. (#4267)
- Changes to CLI syntax for inspect to accept branch option. (#4293)
- Changes to CLI output for list repo and list pipeline to show description. (#4368)
- Changes to CLI output for list commit to show progress and description while removing parent and duration output. (#4368)

## 1.9.8

- Fixes a bug that prevent the `--reprocess` flag in `edit pipeline` from working. (#4232)
- Changes the CLI syntax for `run pipeline` to accept commit branch pairs. (#4262)
- Fixes a bug that caused `pachctl logs --follow` to exit immediately. (#4259)
- Fixes a bug that joins to sometimes miss pairs that matched. (#4256)
- Fixes a bug that prevent pachyderm from deploying on Kuberentes 1.6 without modifying manifests. (#4242)
- Fixes a family of bugs that could cause output and stats commits to remain open and block later jobs. (#4215)

## 1.9.7

- Fixes a bug that prevent pachctl from connecting to clusters with TLS enabled. (#4167)

## 1.9.6

- Fixes a bug which would cause jobs to report success despite datum failures. (#4158)
- Fixes a bug which prevent Disk resource requests in pipelines from working. (#4157)
- Fixes a bug which caused `pachctl fsck --fix` to exit with an error and not complete the fix. (#4155)
- Pachctl contexts now have support for importing Kubernetes contexts. (#4152)
- Fixes a bug which caused Spouts to create invalid provenance. (#4145)
- Fixes a bug which allowed creation, but not deletion, of pipelines with invalid names. (#4133)
- Fixes a bug which caused ListTag to fail with WriteHeader already called. (#4132)
- Increases the max transaction operations and max request bytes values for etcd's deployment. (#4121)
- Fixes a bug that caused `run pipeline` to crash pachd. (#4109)
- Pachctl deploy amazon now exposes several new s3 connection options. (#4107)
- Readds the `--namespace` flag to `port forward`. (#4105)
- Removes and unused field `Batch` from the pipeline spec. (#4104)

## 1.9.5

- Fixes a bug that caused the Salt field to be stripped from restored pipelines. (#4086)
- Fixes a bug that caused datums to fail with `io: read/write on closed pipe`. (#4085)
- Fixes a bug that prevented reading logs from running jobs with stats enabled. (#4083)
- Fixes a bug that prevented putting files into output commits via s3gateway. (#4076)

## 1.9.4

- Fixes a bug (#4053) which made it impossible to read files written to output commits with `put file`. (#4055)
- Adds a flag `--fix` to `pachctl fsck` which will fix some of the issues that it detects. (#4052)
- Fixes a bug (#3879) which caused `pachctl debug dump` to hit max message size issues. (#4015)
- The Microsoft Azure Blob Storage client has been upgraded to the most recent version. (#4000)
- Extract now correctly extracts the `pod_patch` and `pod_spec` for pipelines. (#3964, thanks to @mrene)
- S3Gateway now has support for multi-part uploads. (#3903)
- S3Gateway now has support for multi-deletes. (#4004)
- S3Geteway now has support for auth. (#3937)

## 1.9.3

- Fixes a bug that caused the Azure driver to lock up when there were too many active requests. (#3970)
- Increases the max message size for etcd, this should eliminate errors that would appear with large etcd requests such as those created when deleting repos and pipelines. (#3958)
- Fixes several bugs that would cause commits not to be finished when jobs encountered errors, which would lead to pipelines getting stuck. (#3951)

## 1.9.2

- Fixes a bug that broke Pachyderm on Openshift. (#3935, thanks to @jiangytcn)
- Fixes a bug that caused pachctl to crash when deleting a transaction while no active transaction was set. (#3929)
- Fixes a bug that broke provenance when deleting a repo or pipeline. (#3925)

## 1.9.1

- Pachyderm now uses go modules. (#3870)
- `pachctl diff file` now diffs content, similar to `git diff`. (#3866)
- It's now possible to create spout services as ingress endpoints. (#3829)
- Pachyderm now supports contexts as a way to access multiple clusters. (#3786)
- Fixes a bug that causes `pachctl put file --overwrite` to fail when reading from stdin. (#3882)
- Fixes a bug that caused jobs from run pipeline to succeed when they should fail. (#3872)
- Fixes a bug that caused workers to get stuck in a crashloop. (#3858)
- Fixes a bug that causes pachd to panic when a pipeline had no transform. (#3866)

## 1.9.0

- `pachctl` now has a new, more consistent syntax that's more in line with other container clis such as `kubectl`. (#3617)
- Pachyderm now exposes an s3 interface to the data stored in pfs. (#3411, #3432, #3508)
- Pachyderm now supports transactional PFS operations. (#3658)
- The `--history` flag has been extended to `list job` and `list pipeline` (in addition to `list file`.) (#3692)
- The ancestry syntax for accessing branches (`master^`) has been extended to include forward references i.e. `master.1`. (#3692)
- You can now define service annotations and service type in your pipeline specs. (#3755, thanks to @cfga and @DanielMorales9)
- You can now define error handlers for your pipelines. (#3611)
- Pachyderm has a new command, `fsck` which will check pfs for corruption issues. (#3691)
- Pachyderm has a new command, `run pipeline` which allows you to manually trigger a pipelined on a set of commits. (#3642)
- Commits now store the original branch that they were created on. (#3583)
- Pachyderm now exposes tracing via Jaeger. (#3541)
- Fixes several issues that could lead to object store corruption, particularly on alternative object stores. (#3797)
- Fixes several issues that could cause pipelines to get hung under heavy load. (#3788)
- Fixes an issue that caused jobs downstream from jobs that output nothing to fail. (#3787)
- Fixes a bug that prevent stats from being toggled on after a pipeline had already run. (#3744)
- Fixes a bug that caused `pachctl` to crash in `list commit`. (#3699)
- Fixes a bug that caused provenance to get corrupted on `delete commit`. (#3696)
- A few minor bugs in the output and erroring behavior of `list file` have been fixed. (#3601, #3596)
- Preflight object store tests have been revamped and their error output made less confusing. (#3592)
- A bug that causes stopping a pipeline to create a new job has been fixed. (#3585)
- Fixes a bug that caused pachd to panic if the `input` field of a pipeline was nil. (#3580)
- The performance of `list job` has been greatly improved. (#3557)
- `atom` inputs have been removed and use `pfs` inputs instead. (#3639)
- The `ADDRESS` env var for connecting to pachd has been removed, use `PACHD_ADDRESS` instead. (#3638)

## 1.8.8

- Fixes a bug that caused pipelines to recompute everything when they were restored. (#4079)

## 1.8.7

- Make the 'put file' directory traversal change backwards compatible for legacy branches (#3707)
- Several fixes to provenance (#3734):
    - Force provenance to be transitively closed
    - Propagate all affected branches on deleteCommit
    - Fix weird two branches with one commit bugs
- Added a new fsck utility for PFS (#3734)
- Make stats somewhat toggleable (#3758)
- Example of spouts using kafka (#3752)
- Refactor/fix some of the PFS upload steps (#3750)

## 1.8.6

- The semantics of Cron inputs have changed slightly, each tick will now be a separate file unless the `Overwrite` flag is set to true, which will get you the old behavior. The name of the emitted file is now the timestamp that triggered the cron, rather than a static filename. Pipelines that use cron will need to be updated to work in 1.8.6. See [the docs](https://docs-archive.pachyderm.com/en/v1.8.6/reference/pipeline_spec.html#cron-input) for more info. (#3509)
- 1.8.6 contains alpha support for a new kind of pipeline, spouts, which take no inputs and run continuously outputting (or spouting) data. Documentation and an example of spout usage will be in a future release. (#3531)
- New debug commands have been added to `pachctl` to easily profile running pachyderm clusters. They are `debug-profile` `debug-binary` and `debug-pprof`.  See the docs for these commands for more information. (#3559)
- The performance of `list-job` has been greatly improved. (#3557)
- `pachctl undeploy` now asks for confirmation in all cases. (#3535)
- Logging has been unified and made less verbose. (#3532)
- Bogus output in with `--raw` flags has been removed. (#3523, thanks to @mdaniel)
- Fixes a bug in `list-file --history` that would cause it to fail with too many files. (#3516)
- `pachctl deploy` is more liberal in what it accepts for bucket names. (#3506)
- `pachctl` now respects Kubernetes auth when port-forwarding. (#3504)
- Output repos now report non-zero sizes, the size reported is that of the HEAD commit of the master branch. (#3475)
- Pachyderm will no longer mutate custom image names when there's no registry. (#3487, thanks to @mdaniel)
- Fixes a bug that caused `pod_patch` and `pod_spec` to be reapplied over themselves. (#3484, thanks to @mdaniel)

## 1.8.5
- New shuffle step which should improve the merge performance on certain workloads.

## 1.8.4

- Azure Blob Storage block size has been changed to 4MB due to object body too large errors. (#3464)
- Fixed a bug in `--no-metrics` and `--no-port-forwarding`. (#3462)
- Fixes a bug that caused `list-job` to panic if the `Reason` field was too short. (#3453)

## 1.8.3

- `--push-images` on `create-pipeline` has been replaced with `--build` which builds and pushes docker images. (#3370)
- Fixed a bug that would cause malformed config files to panic pachctl. (#3336)
- Port-forwarding will now happen automatically when commands are run. (#3340)
- Fix bug where `create-pipeline` accepts names which Kubernetes considers invalid. (#3344)
- Fix a bug where put-file would respond `master not found` for an open commit. (#3184)
- Fix a bug where jobs with stats enabled and no datums would never close their stats commit. (#3355)
- Pipelines now reject files paths with utf8 unprintable characters. (#3356)
- Fixed a bug in the Azure driver that caused it to choke on large files. (#3378)
- Fixed a bug that caused pipelines go into a loop and log a lot when they were stopped. (#3397)
- `ADDRESS` has been renamed to `PACHD_ADDRESS` to be less generic. `ADDRESS` will still work for the remainder of the 1.8.x series of releases. (#3415)
- The `pod_spec` field in pipelines has been revamped to use JSON Merge Patch (rfc7386) Additionally, a field, `pod_patch` has been added the the pipeline spec which is similar to `pod_spec` but uses JSON Patches (rfc6902) instead. (#3427)
- Pachyderm developer names should no longer appear in backtraces. (#3436)

## 1.8.2

- Updated support for GPUs (through device plugins).

## 1.8.1

- Adds support for viewing file history via the `--history` flag to `list-file` (#3277, #3299).
- Adds a new job state, `merging` which indicates that a job has finished processing everything and is merging the results together (#3261).
- Fixes a bug that prevented s3 `put-file` from working (#3273).
- `atom` inputs have been renamed to `pfs` inputs. They behave the same, `atom` still works but is deprecated and will be removed in 1.9.0 (#3258).
- Removed `message` and `description` from `put-file`, they don't work with the new multi `put-file` features and weren't commonly used enough to reimplement. For similar functionality use `start-commit` (#3251).

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

* Pipelines now support the “stats” feature.  See the [docs](https://http://docs-archive.pachyderm.com/en/latest/reference/pipeline_spec.html#enable-stats-optional) for details.  (#1998)
* Pipeline cache size is now configurable.  See the [docs](https://docs-archive.pachyderm.com/en/latest/reference/pipeline_spec.html#cache-size-optional) for details.  (#2033)
* `pachctl update-pipeline` now **only** process new input data with the new code; the old input data is not re-processed. If it’s desired that all data are re-processed, use the `--reprocess` flag.  See the [docs](http://docs-archive.pachyderm.com/en/latest/how-tos/updating_pipelines.html) for details.  (#2034)
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

* [Copy elision](http://docs-archive.pachyderm.com/en/latest/managing_pachyderm/data_management.html#shuffling-files) does not work for directories. (#1803)
* Deleting a file in a closed commit fails silently. (#1804)
* Pachyderm has trouble processing large files. (#1819)
* etcd uses an unexpectedly large amount of space. (#1824)
* `pachctl mount` prints lots of benevolent FUSE errors. (#1840)

### New features

* `create-repo` and `create-pipeline` now accept the `--description` flag, which creates the repo/pipeline with a "description" field.  You can then see the description via `inspect-repo/inspect-pipeline`. (#1805)
* Pachyderm now supports garbage collection, i.e. removing data that's no longer referenced anywhere.  See the [docs](http://docs-archive.pachyderm.com/en/latest/managing_pachyderm/data_management.html#garbage-collection) for details. (#1826)
* Pachyderm now has GPU support!  See the [docs](http://docs-archive.pachyderm.com/en/latest/managing_pachyderm/sharing_gpu_resources.html#without-configuration) for details. (#1835)
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
* You can now specify the amount of resources (i.e. CPU & memory) used by your pipelines. (#1683)

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

* Pachyderm now supports auto-scaling: a pipeline's worker pods can be terminated automatically when the pipeline has been idle for a configurable amount of time.  See the `scaleDownThreshold` field of the [pipeline spec](http://docs-archive.pachyderm.com/en/latest/reference/pipeline_spec.html#standby-optional) for details.
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

- A new way to specify how input data is partitioned.  Instead of using two flags `partition` and `incrementality`, we now use a single `glob` pattern.  See the [glob doc](http://docs-archive.pachyderm.com/en/latest/reference/pipeline_spec.html#the-input-glob-pattern) for details.

- Flexible branch management.  In prior versions, branches are fixed, in that a commit always stays on the same branch, and a branch always refers to the same series of commits.  In 1.4, branches are modeled similar to Git's tags; they can be created, deleted, and renamed independently of commits.

- Simplified commit states.  In prior versions, commits can be in many states including `started`, `finished`, `cancelled`, and `archived`.  In particular, `cancelled` and `archived` have confusing semantics that routinely trip up users.  In 1.4, `cancelled` and `archived` have been removed.

- Flexible pipeline updates.  In prior versions, pipeline updates are all-or-nothing.  That is, an updated pipeline either processes all commits from scratch, or it processes only new commits.  In 1.4, it's possible to have the updated pipeline start processing from any given commit.

- Reduced cluster resource consumption.  In prior versions, each Pachyderm job spawns up a Kubernetes job which in turn spawns up N pods, where N is the user-specified parallelism.  In 1.4, all jobs from a pipeline share N pods.  As a result, a cluster running 1.4 will likely spawn up way fewer pods and use fewer resources in total.

- Simplified deployment dependencies.  In prior versions, Pachyderm depends on RethinkDB and etcd to function.  In 1.4, Pachyderm no longer depends on RethinkDB.

- Dynamic volume provisioning.  GCE and AWS users (Azure support is coming soon) no longer have to manually provision persistent volumes for deploying Pachyderm.  `pachctl deploy` is now able to dynamically provision persistent volumes.

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

- Our developer portal is now available at: https://docs.pachyderm.com/latest/
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


# Changelog
## 2.1.8
- red hat push fix 2.1.x - #7626
- Ensure shard ranges contain all paths. - #7620
- Upgrade s2 - #7617
- 2.1.x backport: debug dump: collect logs for all suite=pachyderm pods - #7614
- Added missing default jupyterhub values.yaml - #7575
- [Marketing Ops] Use hyphen in urls - 2.1.x - #7563

## 2.1.7
- Eagerly delete task service group entries - 7552
- Implement compaction cache - 7551
- fix overlooked text circleci matrix adds to job names - 7542
- Initial script for pushing redhat marketplace images - 7535
- migrate ci load test runs off of hub - 7534
- Batch copy operations in unordered writer. - 7531
- Exit finishing early for errored commits. - 7532
- Check for validation errors when choosing a job's base commit. - 7533
- Autoformat the circle CI config and helm values.yaml with Prettier - 7524
- Turn off mockidp role binding when upstream idps - 7515

## 2.1.6
- Added comment on shard size/number thresholds params + dex database name override param - #7509
- Ensure shard ranges are disjoint. - #7502
- Expose configuration for compaction sharding - #7504
- re-enable rootless ci tests - #7481
- temporary circleci release definition - #7487
- add missing --global to git config for build bot - #7491
- fix broken helm chart icon path - #7496
- Add buffering to debug dump stream - #7498
- Decouple content defined chunking from file batching - #7494
- [2.1.x Backport] Add dex secret file support - #7480
- [2.1.x Backport] Expose pachd.identityDatabaseFullNameOverride to configure dex DB name - #7469
- Wordcount example now finds license 13 times - #7468
- add the --wait all for minikube start - #7429
- [2.1.x Backport] Expose PG Bouncer default pool size - #7456
- Block task service setup on etcd client creation. - #7452
- Remove deleted load tests from circle - #7440
- [2.1.x Backport] Make mockIDP user clusterAdmin - #7432
- Distribute load testing - #7430
- [2.1.x Backport] Fix Secure (HTTPS) setting for Minio - #7399
- Helm Publish include --version - #7424
- Fix typo and improve verbiage in pachctl pps cmds	- #7410
- Strip build directory from pachd stack traces	- #7412
- [2.1.x] storage/chunk: Fix TaskChain - #7419
- [2.1.x Backport] Propagating Custom certs referenced in pachd's SSL_CERT_DIR to side-cars - #7417

## 2.1.5
- Expose PG Bouncer default pool size - #7456
- Block task service setup on etcd client creation. - #7452
- Remove deleted load tests from circle - #7440
- Make mockIDP user clusterAdmin - #7432
- Distribute load testing - #7430
- add the --wait all for minikube start - #7429
- Fix Secure (HTTPS) setting for Minio - #7399
- Helm Publish include --version - #7424
- Fix typo and improve verbiage in pachctl pps cmds - #7410
- Strip build directory from pachd stack traces - #7412
- storage/chunk: Fix TaskChain - #7419
- Propagating Custom certs referenced in pachd's SSL_CERT_DIR to side-cars - #7417

## 2.1.4
- Add Snowflake driver - #7406
- Upgrade to latest S2 library - #7398

## 2.1.3
- Fix debug dump segfault on failed jobs - #7377
- Fix: check sql.ErrNoRows in error chain - #7381
- Receive keep alive responses for worker ip renewal - #7378
- Add post upgrade hook for configpod - #7376
- Do not propagate TotalFileSet if parent commit errored in alias commits - #7383

## 2.1.2
- Handling commas in jsonnet argument - #7363
- Run config-job when license is provided via secret - #7348
- Remove init container from enterprise server deployment - #7369
- 
## 2.1.1
- Bug fix for upgrading with more than 10 pipelines - #7360
- Add nodeSelector and annotations to some helm templates - #7354
- Reduce chunk batch size - #7352
- Finer grained locking of Index Writer data - #7350
- Remove pachd's init container - #7347
- Open commits performance changes - #7342
- Allow using Update pipeline to create a pipeline when auth is enabled - #7340

## 2.1.0
Highlights:
- Integration with Structured Data stores, using SQL query pipelines (7248, 7129, 7108, 7035, 7032)
- JSonnet Pipeline Specs (experimental): Pipeline Specs that are scriptable with JSonnet (7154)
- Significantly improved debugging tools:
    - Use Loki for more persistent log collection (6803, 7228, 7271, 7286)
    - Store more information (pipeline history, alias commit information, task information) in debug dumps (7203, 7167, 7312, 7256)
    - Log username of originating user alongside RPCs, for attribution (7231)
- Several new PFS features and performance improvements:
    - Add 'pachctl delete commit' command (alongside the existing 'squash commit') (7094)
    - GC frequency can be configured or disabled via the 'STORAGE_CHUNK_GC_PERIOD' and 'STORAGE_GC_PERIOD' env vars and 'pachd.{storageGCPeriod,storageChunkGCPeriod}' helm values (7284, 7285)
    - Improved caching (7058, 7293), transaction handling (7180, 7308), commit finishing (7047), and GC (7077)
- Significantly improved PPS performance (7040, 7045, 7064, 7066, 7152, 7246, 7260, 7261, 7288, 7120, 7301)
- New Task Service that allows better caching and checkpointing, shared by jobs, compaction, and garbage collection (7143, 7262, 7315)
- Many improvements to the Helm chart:
    - Support deploying a new Pachyderm cluster alongside an existing Enterprise server (7057)
    - Support for nodeSelectors, annotations, and fixed support for tolerations (7082, 7147, 7179, 7183)
    - Also support re-use of existing kubernetes secrets for Pachyderm's Postgres password and OAuth secret (7157, 7192)
    - Add support for HTTP/S proxies (7215)
    - Add 'disableTelemetry' option to disable telemetry in the Pachyderm Console (7207)
    - Add 'pachd.activateAuth' value (true by default) to allow users to activate enterprise without also activating auth on startup (7132)
    - Add 'customCACerts' option, for customers with their own PKI
    - Make Pachyderm's ClusterRole Bindings configurable (7171)

Notable Bug Fixes:
- 7124: RC would not update correctly when a pipeline is deleted and quickly recreated
- 7198: Support use of Kubernetes ingres with the Pachyderm enterprise server
- 7092: Fix 'pachctl mount' panic on unmount, and improve upload performance
- 7296: Fix 'pachctl mount' attempting to commit changes to read-only (output) branches
- 7276: Fix auth error with Cron pipelines
- 7153: Fix incorrect 'meta' branch provenance with services and spouts
- 7114: Block transactions during initialization
- 7227: Fix worker error when deployed alongside an existing postgres instance
- 7164: Show "reason" next to KILLED jobs in 'pachctl list job'
- 7218: Make the kubernetes SecurityContext associated with workers optional, controlled by the pachd.securityContext.enabled helm value
- 7295: Return 404 instead of 400 for paths with trailing slash (fixes Spark jobs)
- 7158: Honor request cancellation in the S3 gateway
- 7250: Remove path for the user container to access Pachyderm's storage secret
## 2.0.8
- Return 404 instead of 400 on trailing slash - #7320
- Obj: Don't hold locks while doing Gets in the cacheClient - #7309
## 2.0.7
- Make Watcher logic more robust, to avoid SubscribeJob missing a job event - #7273
- Adding kubernetes Node Selectors + Tolerations to the helmchart for pachd & etcd - #7177
- To avoid having Cron Pipelines break with Auth, provide Pipelines write access on their Cron input repos - #7279
- Add nil pointer checks during log secret redactions, in case of errors - #7278
## 2.0.6
- Stream lists from batched queries - #7243
- Pass postgres secret reference from pachd to workers - #7235
- Include usernames in gRPC logs - #7239
- Stop pachctl mount panicking on unmount - #7098
- Add loki log collection to debug dump - #7234
- Collect alias commits in debug dump - #7232
## 2.0.5
- Make ingress work for enterprise server - #7224
- Unset Worker SecurityContexts when pachd.securityContext.enabled=false - #7223
- Add http/s proxy settings for pachd - #7222
- CheckStorage rpc and chunk layer integrity checking - #7208
- Do not fail pipeline for transient database issues - #7206
## 2.0.4
- Fix access to console over port-forward. Gate setting of REACT_APP_RUNTIME_SUBSCRIPTIONS_PREFIX by ingress.enabled=false - #7170
- Make Sensitive Helm values injectable via k8s secrets - #7193 #7194 #7188
- Add a helm value - “global.customCaCerts” to load the certs provided in pachd.tls as the root certs for pachd, console, and enterprise-server - #7160
- Support setting pachyderm auth’s cluster role bindings using the helm value: pachd.pachAuthClusterRoleBindings - #7175
- Fix memory leak in s3 gateway due to un-closed connections - #7161
- Fix an issue where deleting a spout would leave behind some data - #7162
- Pipeline specs and job infos from prior versions of a pipeline are now collected in a debug dump - #7169
- Allows the configuration of the Postgresql password via an existing secret - #7176
- Can now configure kubernetes annotations on Deployments, Statefulsets, and ConfigJobs  - #7174 #7189 #7196
- Upgrade Ingress spec from v1beta to v1. Requires Kubernetes v1.19 or higher - #7178
- pachctl list job now shows the reason for killed jobs - #7164
- Config-pod should bootstrap enterprise with communicating with the appropriate port - #7191
## 2.0.3
- Add helm value ingress.uriHttpsProtoOverride to set ingress endpoints with “https” protocol backed by TLS upstream, as opposed to the default TLS configuration using ingress.tls helm values - #7134
- Fix use of passed in helm value: “pachd.oauthRedirectURI“ - #7133
- Require populated PG Passwords on Helm Upgrades - #7131
- Improve large commit finishing performance with autoscaling - #7121
- Add helm value pachd.activateAuth, to control whether the config-job bootstraps auth on the cluster. Defaults to true - #7136
- Fixes a potential pachd crash during initialization - #7114
- Improved the performance of commit finishing and fixed related issues - #7123
- The sharding and concatenating tasks during compaction are now distributed.
- The finishing of commits in different repos is now parallelized.
- Duplicate work that can occur when determining which levels to compact has been removed.
- File sets created through the file set API are now compacted to improve file set read performance.
- Fixes a gRPC leak that can occur when collecting info from workers in pachd.
- Fixes an issue where the job registry would process jobs in the finishing state.
- Fixes a set of deadlocks that can occur in the postgres listener.

## 2.0.2
- Improve tracker GC query performance - #7093
- Add pachctl delete commit command - #7094
- Add the “activateEnterprise” helm value that can be set to bootstrap a cluster’s enterprise/auth features during upgrade releases. During installs, providing the enterpriseLicenseKey will trigger the bootstrap process. - #7104
- Improve logging by disambiguating some commit and job logs - #7111
- Add helm configuration to register a pachyderm cluster with an external enterprise server. #7109
## 2.0.1
- Fixes a goroutine leak that can occur when a modify file operation is canceled - #7055
- Improves performance by reducing the number of postgres requests needed when renewing chunks being uploaded - #7052
- Improves usability by changing helm defaults- activate enterprise automatically and use default localhost URIs when ingress is disabled - #7039
- Improves memory scaling of join and group inputs beyond millions of datums and - - - - Increases logging verbosity by including etcd and pg-bouncer logs in debug dumps - #7053
## 2.0.0 
Introducing Pachyderm 2.0 with several foundational improvements. Read more details [here](https://www.pachyderm.com/blog/getting-ready-for-pachyderm-2/)

### [What’s new](https://docs.pachyderm.com/2.0.x/getting-started/whats-new/)

- New storage architecture and FileSets for better support for small files, content defined chunking for better de-duplication, automatic compression and encryption of chunks, automatic garbage collection, and more
- Pachyderm Enterprise Management to allow site-wide configuration
- New enterprise UI -- Pachyderm Console
- Improved OAuth-based Authentication
- Simplified and unified lineage tracking with Global IDs
- More efficient job run time to significantly reduce job completion time for real world workloads
- Introducing Pachyderm deployment support using Helm Chart
- Updated python-pachyderm 7.0 release which supports Pachyderm 2.x

### Changes from Pachyderm 1.x behavior

- Empty directories are no longer supported
- Default upload behavior changes from append to overwrite
- Full paths in repo must be specified when uploading files
- Automatic file splitting is no longer supported
- `pachctl deploy` has been deprecated in favor of helm charts. See docs for deployment details
- Standby option replaced by autoscaling
- Writing multiple input datums to same output file is no longer supported and will result in an error 

Updated docs are at [docs.pachyderm.com](https://docs.pachyderm.com/latest/)

Try Pachyderm 2.0 on [hub.pachyderm.com](https://hub.pachyderm.com/landing?redirect=%2F)

Pachyderm 2.x is not backwards compatible with Pachyderm 1.x data formats. If you require assistance or have any questions, please contact [support@pachyderm.com](mailto:support@pachyderm.com)

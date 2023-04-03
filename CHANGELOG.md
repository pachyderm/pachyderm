
# Changelog

## 2.5.3
* Basic PPS UI  https://github.com/pachyderm/pachyderm/pull/8557
* Return error instead of deleting a project containing repos https://github.com/pachyderm/pachyderm/pull/8631
* Correct context for processPutFileTask in pfload worker https://github.com/pachyderm/pachyderm/pull/8638
* Add a warning message when deleting the current project https://github.com/pachyderm/pachyderm/pull/8637
* Correct project-already-exists error message https://github.com/pachyderm/pachyderm/pull/8628
* Update Go version to 1.20.2 https://github.com/pachyderm/pachyderm/pull/8640
* Djanicek/core 1481/debug fix https://github.com/pachyderm/pachyderm/pull/8639

## 2.5.2

* Djanicek/core 1481/debug fix https://github.com/pachyderm/pachyderm/pull/8646
* back port new nightly load iac envshttps://github.com/pachyderm/pachyderm/pull/8654
* bump console version to 2.5.2 https://github.com/pachyderm/pachyderm/pull/8657
## 2.5.1 

* [2.5.x Backport][CORE-1463] pgbouncer pg_isready https://github.com/pachyderm/pachyderm/pull/8630
* bump console version to 2.5.1-1 https://github.com/pachyderm/pachyderm/pull/8632


## 2.5.0
### Projects
- Add Projects as a Pachyderm resource - #8078
- Create migration for default project - #8108
- Link projects to resources - #8114
- Add current project to pachctl config context - #8150
- Make PPS pipeline inspection project-aware - #8154
- Update data structures - #8156
- Update PFS client library to support projects - #8159
- Update PPS client library to be project-aware - #8191
- Pipeline lifecycle - #8272
- Make repo role bindings aware of projects - #8272
- scope 'pachctl draw' to projects - #8536
- list project active marker - #8461
- diff file should apply --project to both files unless --old-project is set - #8474
- include project field in 'pachctl list job <ID>' & in 'pachctl inspect job' - #8462
- project names must start with an alphanumeric character - #8450
- Make `pachctl copy file` able to copy files between projects - #8507
- Add `pachctl auth check|set project` - #8467
- Make `pachctl list pipelines` filter on projects - #8446
- Migrate pipeline user names to be aware of default project - #8442
- Make `pachctl auth check|get|set repo` aware of current project - #8431
- Add PROJECT column to PFS pprinter - #8433
- Remove ProjectViewer and ProjectWriter - #8419
- Authorize DeleteProject - #8410
- ListRepo - filter on projects and check permissions - #8382
- Create role bindings for projects created before auth activation - #8409
- Make ModifyRoleBinding project aware - #8354
- Make ListJob project-aware - #8211
- Extend auth with project primitives, and implement CreateProject - #8235
- Do not delete projects with repos or pipelines unless forced - #8526
- Update ListJobSetRequest and ListJobRequest to take Project messages - #8477
- Disallow changing project or pipeline name in pachctl edit pipeline - #8465
- Add project name to AddSpanToAnyExisting calls - #8466
- Change ListRepo to take a list of Project messages rather than strings - #8435
- Correct pachctl delete repo --all for projects - #8393
- Check that project exists before creating repo - #8448
- Project-aware pachctl list commit - #8447
- Update debug service to be project-aware - #8376
- Support for projects in extended traces - #8390
- Update S3 sidecar to be project-aware - #8370
- Default project in transactional repo creation - #8380
- Plumb contextual project through pachctl commands - #8348
- S3 gateway and projects - #8365
- Update FUSE mounts for projects - #8311
- Add short flag -A for --all-projects - #8561
- Address the issue of overlong Kubernetes resource names - #8559
- Address getLogsLoki project TODO - #8553

### Pachctl
- ListJobSet pagination - #8351
- ListJob pagination - #8350		
- WalkFile pagination - #8475	
- Add Pachctl connect command - #8514	
- support tgz extension for pachctl put file untar - #8361
- Add kubernetes event logs to pachctl - #8518

### Monitoring/Logging
- Add kubernetes event logs to pachctl - #8518
- re-introduce explicit audit log - #8540
- log http error logs at Debug level - #8470
- paused: scaleDownWorkers: supply objectmeta; capture all errors - #8508
- Debug dump improvements; take 2 - #8487
- Some adjustments to logging - #8471
- Implement the `log` package - #8411
- dlock: add logging around lock acquisition and release - #8469
- script to benchmark putfile - #8404
- envoy: switch to JSON; log listener events - #8398
- Adjust envoy default stream window to increase upload throughput - #8397

### Performance Improvements		
- Modify GetFileURL Coordination Algorithm to Batch Files into Tasks by Size - #8544
- Fix setting environment variables for non s3 storage backends - #8552
- Update PutFileURL Task Model to Utilize PathRanges and Offsets  - #8520
- Enable 'inSidecars' in pachw by default - #8543
- Add support for affinity and tolerations - #8537
- Migrate PutFileURL and GetFileURL to Go-Cloud-SDK - #8506
- Bugfix for Pachw - #8483
- Add Helm Configuration for Storage Parameters - #8458
- Run Pachd in 'PachW' Mode and Autoscale To Handle Tasks - #8424
- Deal with subprocesses cleanly in the worker - #8385
- Datum batching - #8443

### Extensions
- Projects support - #8459
- Add project filter to UI - #8523
- Show message if datum cycler hidden - #8511
- Qualify unmounted repos keys by project - #8515
- Projects support - #8415
- Prevent mounting non-existent branch in a read-only mount - #8341

#### Bugs/Fixes
- GetPermissionsResponse.Permissions should respect the requested resource type - #8460
- Use the correct SQL to update auth tokens in migration - #8513
- Ensure ListJobSet is actually returning the filtered list of jobs - #8484
- Fix pachctl auth get --help message - #8413
- Fix the locking in the consistent hashing package - #8502
- Ensure we only create jobs for output branch commits - #8486
- Fix several goroutine leaks when closing a service environment - #8476
- Scale down pipeline workers when finishing a commit - #8464
- Distributed put / get file url - #8394
- Revert "Test tweaking default shard threshold - #8230)" - #8355
- Tweak default shard threshold - #8230
- Don't bail if the cluster deployment ID of the cluster doesn't match - #8529
- envoy: ignore 200s served to kube-probe - #8402
- Allow pipeline spec to directly specify k8s tolerations - #8517
- helm: enable proxy by default; add NOTE about externalService deprecation - #8434
- proxy: allow helm chart to set tolerations and node selector; don't mount service account token - #8279
- worker_rc: set default requests if resource_requests and resource_limits are both empty - #8386
- oidc: avoid panicking on invalid input - #8387
- Implement the S3 Gateway bucket name proposal - #8516
- Update version of jaeger all-in-one image - #8425
- re-add pool size for pgbouncer image - #8345
- Don't print warning about default login config when a real ID provider is configured for helm deployment - #8535
- More accurate stop pipeline message - #8389
- Correct nil pointer dereference from ListFile - #8422
- Fix bad branch not found message - #8416
- Make DeleteRepos and DeletePipelines similar - #8510
- schema migration to extend length limit of auth_token subjects - #8580
- Don't log clients that don't send a version number - #8582
- helm: allow loki internal comms to use large message sizes; necessary… - #8589
- Propagate pachw in sidecar variable to sidecars. - #8599
- concurrency options - #8601
- handle different formats of loki logs for kube events - #8603
- list datum fix - #8610
- Fix datum logger - #8612

## 2.4.6

* [2.4.x Backport] [CORE-1463] switch pgbouncer liveness to pg_isready (#8578) by @tybritten in https://github.com/pachyderm/pachyderm/pull/8584
* bump console version 2.4.6 by @molinamelendezj in https://github.com/pachyderm/pachyderm/pull/8587
* bump jupyter-extension-build python cache by @molinamelendezj in https://github.com/pachyderm/pachyderm/pull/8588
## 2.4.5

* [PFS-46] Fix directory creation with path range requests (2.4.x) by @brycemcanally in https://github.com/pachyderm/pachyderm/pull/8545
* bump console version 2.4.5-1 by @molinamelendezj in https://github.com/pachyderm/pachyderm/pull/8560

## 2.4.4

* cycle new fingerprints (#8504) by @molinamelendezj in https://github.com/pachyderm/pachyderm/pull/8505
* [CORE-1441] [2.4.x] leave pipelines in STANDBY until they're FINISHED by @msteffen in https://github.com/pachyderm/pachyderm/pull/8500
* [2.4.x] Warn about outdated pachctl (#8478) by @jrockway in https://github.com/pachyderm/pachyderm/pull/8492
* [CORE-1391] Make a 'pachyderm-tap/pachctl' formula represent the latest stable release (#8473) by @acohen4 in https://github.com/pachyderm/pachyderm/pull/8519
* bump console version by @molinamelendezj in https://github.com/pachyderm/pachyderm/pull/8521

## 2.4.3 

* [2.4.x port][CORE-1231] pass pachctl args to server for audit logging by @armaanv in https://github.com/pachyderm/pachyderm/pull/8423
* security: update http2 (#8428) by @jrockway in https://github.com/pachyderm/pachyderm/pull/8432
* [Backport 2.4.x] Add timeout to queryLoki (#8449) by @albscui in https://github.com/pachyderm/pachyderm/pull/8451
* s3: log net/http errors at Debug level by @jrockway in https://github.com/pachyderm/pachyderm/pull/8456
* [Backport 2.4.x] Add number of tokens revoked to RevokeToken response (#8453) by @albscui in https://github.com/pachyderm/pachyderm/pull/8457
* [2.4.x backport][Jupyter] Fix datums-related error message when notebooks starts up by @smalyala in https://github.com/pachyderm/pachyderm/pull/8479
* [2.4.x] Increase the reliability of debug dumps by @jrockway in https://github.com/pachyderm/pachyderm/pull/8488
* [2.4.x Backport] Gracefully handle old pipelines where output commit is ALIAS and metacommit is AUTO (#8485) by @acohen4 in https://github.com/pachyderm/pachyderm/pull/8493
* bump console version by @djanicekpach in https://github.com/pachyderm/pachyderm/pull/8494
* 2.4.x update release test by @djanicekpach in https://github.com/pachyderm/pachyderm/pull/8495
* disable cgo in pachctl by @djanicekpach in https://github.com/pachyderm/pachyderm/pull/8498
## 2.4.2

* [PFS-10] Fix Misleading Version Check Logic 2.4.x Backport by @FahadBSyed in https://github.com/pachyderm/pachyderm/pull/8399
* [2.4.x Backport] Remove default controlplane tolerations from promtail (#8396) by @tybritten in https://github.com/pachyderm/pachyderm/pull/8406
* [2.4.x backport] s3gateway: remove limits on file uploads by @jrockway in https://github.com/pachyderm/pachyderm/pull/8420
## 2.4.1

* [2.4.x backport][Mount Server] Mount latest non-alias commit on branch by @smalyala in https://github.com/pachyderm/pachyderm/pull/8366
* Core 1123 2.4.x Backport by @FahadBSyed in https://github.com/pachyderm/pachyderm/pull/8384
## 2.4.0

## Core Worker & `pachd` Improvements
- Enable Memory-Backed Pachyderm Worker Volumes - #8262
- Fix worker keepalive timeout - #8246
- Disable S3 server when running in paused mode - #8168
- regenerate protos as they've gone out of sync - #8166
- Factor out S3 server	- #8019
- Handle signals uniformly	- #8033
- Use service environment interface - #8027
- Plumb a context through cmdutil.Main	- #8013
- Start sidecar PFS server in sidecar mode - #8124

## Logging
- loki timeout and line num change - #8329
- Remove error wrap checking from some internal interfaces.	- #7447
- Update inspectPipeline error message for non-existent pipelines	- #8101
- add additional buildinfo to version	- #8034
- Use errgroups to manage goroutines in pachd	- #8020
- Factor out Prometheus server	- #8018
- grpc: log request ID	- #8222

## `pachctl` Improvements 
- Draw pach DAGs on the command line	- #7304
- List pipelines at a commit set	- #8221
- Fix issue with pretty printing JobInfo.Started	- #8135
- pachctl: add "buildinfo" command - #8031

## Pagabale API
- add startedTime marker for ListCommit - #8174
- ListFile pagination - #8335
- ListDatum pagination- #8336

## Performance 
- Propagate shard config to pachd sidecar - #8369
- changes compactionShard settings to str - #8368
- Fix input reordering with cross inputs - #8357
- metadata and data sharding configurations - #8363
- Fix file set renewal in new compaction algorithm	- #8227
- Cache Symlink Commit Uploads	- #8172
- Implement pachctl put file untar	- #8167
- Implement PPS datum sharding	- #7851
- Support Locking Repos Across Multiple PFS Masters Using Consistent Hashing Library - #8045
- Implement chunk prefetcher - #8200

## Proxy 
- Fix file set renewal in new compaction algorithm - #8227

## Security
- Fix list datum input with auth - #8359
- Cache Symlink Commit Uploads	- #8172
- Implement pachctl put file untar	- #8167
- Implement PPS datum sharding	- #7851
- Support Locking Repos Across Multiple PFS Masters Using Consistent Hashing Library - #8045	

## Snowflake Integration
- Implement chunk prefetcher - #8200	

## New Contributors
* @Juneezee made their first contribution in https://github.com/pachyderm/pachyderm/pull/7660
* @harrisonfang made their first contribution in https://github.com/pachyderm/pachyderm/pull/8058

**Full Changelog**: https://github.com/pachyderm/pachyderm/compare/v2.1.0...v2.4.0

## 2.3.9

* Migration load testing (2.3.x) by @brycemcanally in https://github.com/pachyderm/pachyderm/pull/8343

## 2.3.8 

* Backport [CORE-1198] Add a timeout to loki requests from debug dump (#8329) by @armaanv in https://github.com/pachyderm/pachyderm/pull/8331
## 2.3.7

* Backport CORE-1123 to Pachyderm 2.3 by @FahadBSyed in https://github.com/pachyderm/pachyderm/pull/8283
* Backport Fix for Spout pipelines not able to restart by @albscui in https://github.com/pachyderm/pachyderm/pull/8310
* bump console version 2.3.7 by @molinamelendezj in https://github.com/pachyderm/pachyderm/pull/8326
* try temp removing python cache by @molinamelendezj in https://github.com/pachyderm/pachyderm/pull/8328
## 2.3.6 

* [2.3.x backport] update etcd to an image that has an arm64 build (#8169) by @jrockway in https://github.com/pachyderm/pachyderm/pull/8224
* Add extension build as a dependency to publish by @chainlink in https://github.com/pachyderm/pachyderm/pull/8250
* [2.3.x] Jupyter cache bump by @chainlink in https://github.com/pachyderm/pachyderm/pull/8248
* [2.3.x backport] PPS: Be more careful about k8s errors during RC updates by @jrockway in https://github.com/pachyderm/pachyderm/pull/8264
* [2.3.x] Proxy infer issuerURI by @seslattery in https://github.com/pachyderm/pachyderm/pull/8271
* [2.3.x] require scheme on userAccessibleOauthIssuerHost by @seslattery in https://github.com/pachyderm/pachyderm/pull/8275
* bump console version by @molinamelendezj in https://github.com/pachyderm/pachyderm/pull/8274
## 2.3.5 

* [2.3.x Backport] Use external postgres secret when requested (#8138) by @acohen4 in https://github.com/pachyderm/pachyderm/pull/8215
## 2.3.4

* [2.3.x] Don't bother deleting stack if one does not exist, set correct working directory  by @chainlink in https://github.com/pachyderm/pachyderm/pull/8207
* [2.3.x] Install current version of pach in jupyter CI by @chainlink in https://github.com/pachyderm/pachyderm/pull/8204
* bump console version by @molinamelendezj in https://github.com/pachyderm/pachyderm/pull/8217
## 2.3.3

* [2.3.x Backport] liveness probe to not use URI (#8128) by @tybritten in https://github.com/pachyderm/pachyderm/pull/8155
* Add govulncheck to CI, fix vulnerable dependency (#8147) by @jrockway in https://github.com/pachyderm/pachyderm/pull/8152
* [2.3.x] upgrade loki-stack (#8092) by @jrockway in https://github.com/pachyderm/pachyderm/pull/8141
* bump console version by @molinamelendezj in https://github.com/pachyderm/pachyderm/pull/8178
* Jupyter merge 23x backport by @chainlink in https://github.com/pachyderm/pachyderm/pull/8177
* images needed for tests (#8185) by @molinamelendezj in https://github.com/pachyderm/pachyderm/pull/8186
* updated base branch in config by @chainlink in https://github.com/pachyderm/pachyderm/pull/8184
* Don't run preview on 2.3.x branch by @chainlink in https://github.com/pachyderm/pachyderm/pull/8187
* Shell check fix for config.sh by @chainlink in https://github.com/pachyderm/pachyderm/pull/8190

## 2.3.2

* [2.3.x] Enable IAM Login for Cloud SQL Auth Proxy by @BOsterbuhr in https://github.com/pachyderm/pachyderm/pull/8086
* update image path (#8104) by @molinamelendezj in https://github.com/pachyderm/pachyderm/pull/8106
* Reintroduce UPGRADE_NO_OP (#8113) by @acohen4 in https://github.com/pachyderm/pachyderm/pull/8118
* fix circle ci nightly load test jobs (#8117) by @molinamelendezj in https://github.com/pachyderm/pachyderm/pull/8119
* [CORE-824] Disallow configuring logs in text format by @FahadBSyed in https://github.com/pachyderm/pachyderm/pull/8112
* [CORE-448] Spread alias commits from output branches to meta branches… by @acohen4 in https://github.com/pachyderm/pachyderm/pull/8116
* Spark compatibility in the S3 gateway by @lukemarsden in https://github.com/pachyderm/pachyderm/pull/8115
* add build-docker-image dependency for load tests (#8127) by @molinamelendezj in https://github.com/pachyderm/pachyderm/pull/8131
* Add mount server to GH release by @chainlink in https://github.com/pachyderm/pachyderm/pull/8132
* build mount server for distribution in release workflow (#8136) by @molinamelendezj in https://github.com/pachyderm/pachyderm/pull/8137

## 2.3.1

- Remove Loki storageClassName default by - #8089
- helm: make upgrades idempotent with respect to the Values dictionary - #8081
- Fix service pipelines not picking up new jobs. - #8076
- Adds Proxy.host   - #8074
- use localhostIssuer=true even when network routing allows it to be false - #8070
- Raise kubeEventTail mem limit default to 100Mi  - #8051
- Go 1.19 and ioutil refactor - #8054
- Set --platform for docker builds; backport go version improvements - #8046
- PG Bouncer Liveness Probe fix  - #8043
- release: integrate manifests into the release process - #8040
- build info - #8037
- Backport ARM64 builds - #8032
- Disable Automount at the ServiceAccount Level - #7993
- Parallel Deploy Tests - #8003
- Filter ListDatum results - #7986
- Configure Additional Dex Clients via Helm  - #8002
- Deprecate enterprise root token and activateEnterprise helm values  - #7962
- Fix user configured resources for pgbouncer deployment - #7997
- Add errcheck to golangci-lint - #7944
- fix MatchInvertedFail test - #7989
- Prevent multiple lokis from mixing logs - #7983
- Implement PPS load test DAG functionality - #7807
- Make logs clearer - #7969
- Disable service account token automounting for loki by default - #7972
- Support YAML Connector Configs - #7872
- Improve file iteration for debug analyze. - #7954
- Fix License Count Error - #7951
- use the right 'context' - #7950
- Append custom Cluster RBAC to mock RBAC - #7952
- fix userAccessibleOauthIssuerHost - #7947
- added data-clickid tracking - #7935
- Update more dependencies in spout example - #7949
- Update deps & increment Go version in Spouts example - #7946
- Add SSL config to listener config - #7756
- Update insecure dependencies - #7932
- Add command to serve pach API from a debug dump - #7926
- New SQL ingest pipeline architecture design - #7900
- Fix lint and upgrade to go1.18.4 - #7928
- Console should restart on helm upgrades - #7925
- Remove liveness probes for pachd - #7894
- Rewrite Gatk make target to more closely match the Gatk README.md - #7903
- postgres: hard-code a known password; allowing upgrades without specifying a password - #7922
- worker_rc: make the created service headless - #7877
- Support 63-character pipeline names - #7897
- Re-enable S3 tests. - #7906
- Fixes/adds commit mounting. - #7673
- Add runAsUser/Group to all containers - #7893
- Consolidate the kube-event-tail and pgbouncer images. - #7895
- Fix copy url params - #7890
- alpha.0 fixes - #7884
- Keep only Pachyderm logs - #7882
- Bootstrap Pachyderm before serving external traffic  - #7879
- Give the internal auth user cluster admin perms  - #7885
- Reduce the default cluster pool to 3 - #7886
- Fix Update OIDC Client in Bootstrap - #7862
- Log processing of tasks - #7850
- bump ci release job to 1.18 - #7859
- Add helm templates and config for kube-event-tail - #7847
- Register one service instance with both the Internal + External GRPC Servers  - #7846
- Add zombie data detection option to fsck. - #7749
- Continue to use pachd.oauthClientSecretSecretName - #7857
- helm: add a startup probe for pachd - #7860
- update go 1.18 - #7855
- Remove explicit OauthClientSecret decl from test infra - #7852
- Continue deduplicating main.go - #7845
- Add codecov for unit tests. Separate out unit tests using buildtags - #7763
- Remove refs to pachtf docker image - #7849
- Enable Loki by default - #7759
- Distinguish transaction metrics with ErrTxDone. - #7842
- Start deduplicating code in main.go - #7841
- Add missing probes and resources - #7828
- Remove hardcoded images and move to values file - #7827
- [Mount Server] Read from config file on startup - #7836
- Transactionify Auth activation in EnvBootstrap - #7817
- Look for upstream-idps as secret key - #7840
- Update Upgrade test versions - #7839
- Add List() to ReadWrite Postgres Collections - #7835
- add oidc health check to envoy - #7834
- Reorganize load tests and add new load tests - #7703
- env bootstrap enterprise member cluster - #7833
- Add a simple log wrapper to Loki stream entries - #7823
- envoy: use the default idle_timeout - #7765
- Remove pachtf Docker image - #7830
- worker: driver: improve error handling around /etc/passwd and /etc/group inspection - #7831
- priority class names - #7770
- Add comma to json - #7829
- Configurable Security Contexts - #7745
- make EnvBootstrap interface consistent - #7818
- Bootstrap Embedded Pach Cluster's Auth via environment - #7804
- [Mount Server] Upload mount server binary for each commit - #7790
- circle: make run_tests set pipefail - #7811
- Index cache - #7789
- Few changes to reduce per file overhead - #7796
- Add code to dump a database using pagination - #7786
- Fix chomp in pachyderm-bootstrap-config - #7809
- NewAuthServer should include EnvBootstrap - #7771
- Fix TestEnterpriseServerMember - #7797
- Fix Wikipedia test - #7800
- Mount server - API to support refresh button in Jupyter extension - #7788
- Deflake TestUpdatePipelineRunningJob - #7781
- Mount server separate binary - #7747
- Avoid scrubbing get file error message. - #7782
- Pachctl ARM64 - #7305
- Fix IDP Connectors Gitops upgrade - #7774
- Decouple minikubetestenv address from pachctl config - #7773
- [Mount Server] Add endpoint to check if mount server is running - #7767
- Add basic postgres stats to pachctl debug dump - #7761
- Pure Refactor - Reordering in preparation for more bootstrapping - #7772
- makefile: protos: build proto compilation container for the host - #7766
- Let Envoy terminate TLS - #7664
- Clean glob in multiple places. - #7754
- Bootstrap Identity Service Via environment - #7755
- driver: avoid an unlikely-to-succeed call to setgroups() when running non-root user code - #7757
- Add Enterprise Server/Member deploy test - #7744
- Expose whether to write header via pachtf - #7742
- [Mount Server] API modifications for sidecar change - #7693
- Check status code when importing jsonnet. - #7728
- Use load balancer on windows - #7727
- add a http handler for oidc health check - #7634
- Consult RC when un-crashing pipeline. - #7717
- Create enterprise secret when a pachd.enterpriseSecret is provided - #7710
- Fix Enterprise Gitops deployment - #7709
- Sensible enterprise dev values.yaml - #7708
- Bootstrap License Server via environment - #7679
- refactor: replace strings.Replace with strings.ReplaceAll - #7660
- [Mount Server] Add logic for repo access - #7653
- ActivateAuth with a root token injected via pachd's environment - #7654
- Adds a churn example. - #7588
- Make inputs failed message clearer - #7558
- Split out PPS auth tests - #7652
- Surface missing secret - #7628
- Mount server commit - #7632
- Capture docker image SHA for datums - #7587
- add retry to egress - #7610
- Capture the relevant error in test retry loop. - #7615
- Show total input datums rather than running total - #7581

## 2.3.0 
## What's Changed
- helm: make upgrades idempotent with respect to the Values dictionary - #8081
- Fix service pipelines not picking up new jobs. - #8076
- Adds Proxy.host   - #8074
- use localhostIssuer=true even when network routing allows it to be false - #8070
- Raise kubeEventTail mem limit default to 100Mi  - #8051
- Go 1.19 and ioutil refactor - #8054
- Set --platform for docker builds; backport go version improvements - #8046
- PG Bouncer Liveness Probe fix  - #8043
- release: integrate manifests into the release process - #8040
- build info - #8037
- Backport ARM64 builds - #8032
- Disable Automount at the ServiceAccount Level - #7993
- Parallel Deploy Tests - #8003
- Filter ListDatum results - #7986
- Configure Additional Dex Clients via Helm  - #8002
- Deprecate enterprise root token and activateEnterprise helm values  - #7962
- Fix user configured resources for pgbouncer deployment - #7997
- Add errcheck to golangci-lint - #7944
- fix MatchInvertedFail test - #7989
- Prevent multiple lokis from mixing logs - #7983
- Implement PPS load test DAG functionality - #7807
- Make logs clearer - #7969
- Disable service account token automounting for loki by default - #7972
- Support YAML Connector Configs - #7872
- Improve file iteration for debug analyze. - #7954
- Fix License Count Error - #7951
- use the right 'context' - #7950
- Append custom Cluster RBAC to mock RBAC - #7952
- fix userAccessibleOauthIssuerHost - #7947
- added data-clickid tracking - #7935
- Update more dependencies in spout example - #7949
- Update deps & increment Go version in Spouts example - #7946
- Add SSL config to listener config - #7756
- Update insecure dependencies - #7932
- Add command to serve pach API from a debug dump - #7926
- New SQL ingest pipeline architecture design - #7900
- Fix lint and upgrade to go1.18.4 - #7928
- Console should restart on helm upgrades - #7925
- Remove liveness probes for pachd - #7894
- Rewrite Gatk make target to more closely match the Gatk README.md - #7903
- postgres: hard-code a known password; allowing upgrades without specifying a password - #7922
- worker_rc: make the created service headless - #7877
- Support 63-character pipeline names - #7897
- Re-enable S3 tests. - #7906
- Fixes/adds commit mounting. - #7673
- Add runAsUser/Group to all containers - #7893
- Consolidate the kube-event-tail and pgbouncer images. - #7895
- Fix copy url params - #7890
- alpha.0 fixes - #7884
- Keep only Pachyderm logs - #7882
- Bootstrap Pachyderm before serving external traffic  - #7879
- Give the internal auth user cluster admin perms  - #7885
- Reduce the default cluster pool to 3 - #7886
- Fix Update OIDC Client in Bootstrap - #7862
- Log processing of tasks - #7850
- bump ci release job to 1.18 - #7859
- Add helm templates and config for kube-event-tail - #7847
- Register one service instance with both the Internal + External GRPC Servers  - #7846
- Add zombie data detection option to fsck. - #7749
- Continue to use pachd.oauthClientSecretSecretName - #7857
- helm: add a startup probe for pachd - #7860
- update go 1.18 - #7855
- Remove explicit OauthClientSecret decl from test infra - #7852
- Continue deduplicating main.go - #7845
- Add codecov for unit tests. Separate out unit tests using buildtags - #7763
- Remove refs to pachtf docker image - #7849
- Enable Loki by default - #7759
- Distinguish transaction metrics with ErrTxDone. - #7842
- Start deduplicating code in main.go - #7841
- Add missing probes and resources - #7828
- Remove hardcoded images and move to values file - #7827
- [Mount Server] Read from config file on startup - #7836
- Transactionify Auth activation in EnvBootstrap - #7817
- Look for upstream-idps as secret key - #7840
- Update Upgrade test versions - #7839
- Add List() to ReadWrite Postgres Collections - #7835
- add oidc health check to envoy - #7834
- Reorganize load tests and add new load tests - #7703
- env bootstrap enterprise member cluster - #7833
- Add a simple log wrapper to Loki stream entries - #7823
- envoy: use the default idle_timeout - #7765
- Remove pachtf Docker image - #7830
- worker: driver: improve error handling around /etc/passwd and /etc/group inspection - #7831
- priority class names - #7770
- Add comma to json - #7829
- Configurable Security Contexts - #7745
- make EnvBootstrap interface consistent - #7818
- Bootstrap Embedded Pach Cluster's Auth via environment - #7804
- [Mount Server] Upload mount server binary for each commit - #7790
- circle: make run_tests set pipefail - #7811
- Index cache - #7789
- Few changes to reduce per file overhead - #7796
- Add code to dump a database using pagination - #7786
- Fix chomp in pachyderm-bootstrap-config - #7809
- NewAuthServer should include EnvBootstrap - #7771
- Fix TestEnterpriseServerMember - #7797
- Fix Wikipedia test - #7800
- Mount server - API to support refresh button in Jupyter extension - #7788
- Deflake TestUpdatePipelineRunningJob - #7781
- Mount server separate binary - #7747
- Avoid scrubbing get file error message. - #7782
- Pachctl ARM64 - #7305
- Fix IDP Connectors Gitops upgrade - #7774
- Decouple minikubetestenv address from pachctl config - #7773
- [Mount Server] Add endpoint to check if mount server is running - #7767
- Add basic postgres stats to pachctl debug dump - #7761
- Pure Refactor - Reordering in preparation for more bootstrapping - #7772
- makefile: protos: build proto compilation container for the host - #7766
- Let Envoy terminate TLS - #7664
- Clean glob in multiple places. - #7754
- Bootstrap Identity Service Via environment - #7755
- driver: avoid an unlikely-to-succeed call to setgroups() when running non-root user code - #7757
- Add Enterprise Server/Member deploy test - #7744
- Expose whether to write header via pachtf - #7742
- [Mount Server] API modifications for sidecar change - #7693
- Check status code when importing jsonnet. - #7728
- Use load balancer on windows - #7727
- add a http handler for oidc health check - #7634
- Consult RC when un-crashing pipeline. - #7717
- Create enterprise secret when a pachd.enterpriseSecret is provided - #7710
- Fix Enterprise Gitops deployment - #7709
- Sensible enterprise dev values.yaml - #7708
- Bootstrap License Server via environment - #7679
- refactor: replace strings.Replace with strings.ReplaceAll - #7660
- [Mount Server] Add logic for repo access - #7653
- ActivateAuth with a root token injected via pachd's environment - #7654
- Adds a churn example. - #7588
- Make inputs failed message clearer - #7558
- Split out PPS auth tests - #7652
- Surface missing secret - #7628
- Mount server commit - #7632
- Capture docker image SHA for datums - #7587
- add retry to egress - #7610
- Capture the relevant error in test retry loop. - #7615
- Show total input datums rather than running total - #7581
  
## 2.2.7
- Don’t segfault on nil field - #8001
- Update RCs unconditionally. - #7995
## 2.2.6
- Fix swallowed error in PFS DeleteFile - #7921
- snowflake fixes - #7892

## 2.2.5
- Fix upgrade for 2.2.x to 2.2.3+ without console #7874

## 2.2.4
- Remove Log Parsing in lokiutil - #7863

## 2.2.3

- Enable console by default - #7813
- Enterprise Members shouldn't run the Identity/Dex Service - #7785
- Fix cluster perms getting reset by config-pod - #7806
- Add 'pachctl auth revoke' - #7794
- Fix Etcd annotations indent - #7802
- Fix quotes around Enterprise Server address in helm values - #7798
- Remove redundant load test - #7792
- Pass PPS Egress URL correctly - #7791
- Move nightly load tests to pulumi envs - #7780
- Bump load test env creation timeout - #7784
- Enable TLS support for Postgres - #7607
- Add logging response transformer to omit activation code from GetState - #7762
- Improve Loki defaults and Upgrade loki-stack - #7760
- Support JSON logs, for GCP - #7683

## 2.2.2

- Add Support for Plurals in Pachctl - #7751
- egress of CSVs with headers	- #7746
- Various Loki improvements	- #7733
- avoid using 31800 for all nodeports, give each test a unique port - #7735
- Use a unique serve mux per port, allow health checks on OIDC/Identity ports - #7730
- CSV header - #7721
- Match helm appVersion to version. - #7715
- Always prep test results for upload to circle - #7711
- Enterprise Server Cmd Tests against test cluster pachd instances - #7705
- Make deletion progress more visible. - #7706
- Fuse Tests shouldn't depend on 'default' pachd instance - #7700
- Deploy Loki  in test clusters - #7699

## 2.2.1

- Include gRPC status codes for more errors. - #7680
- Enable IAM Login for Cloud SQL Auth Proxy - #7678
- Migrate Cmd Tests to Client Factory - #7672
- Fix a flaky etcd collection test, improve etcd logging - #7676
- Simple put-file progress indication - #7677
- Eagerly delete task service group entries - #7645
- Cherry-pick Snowflake VARIANT support = #7661
- Include more pod failure reasons - #7651
- Worker Master deletes meta files for unskippable datums - #7642
- Renew temporary filesets produced in handleDatumSet task - #7646
- Fix deployment test - #7644
## 2.2.0

- [CORE-586] Update how we handle null values in CSV  - #7623	
- Ensure shard ranges contain all paths.  - #7621		
- Upgrade s2  - #7616	
- 2.2.x backport: debug dump: collect logs for all suite=pachyderm pods  - #7613		
- [CORE-572] Enable copying root directory of source repo  - #7606	
- Fix annotations setting in pachd (2.2.x merge)  - #7602		
- 2.2.x patch - Convert SQL numeric types to string  - #7603		
- [CORE-550] Pretty-print egress  - #7598		
- CORE-558 rename "k8s_secret" to "name" in Egress secret  - #7580		
- [CORE-370] Fix reference to non-existent AccessLevel  - #7582		
- CORE-387 New Egress API  - #7522		
- [CORE-295] Check repo existence before put file  - #7549		
- Make postgres watches respect contexts. - #7544		
- Collect all minikubetestenv logs, avoid minikubetestenv port conflicts - #7550		
- Implement compaction cache - #7538		
- Migrate Auth tests to Client Factory - #7546		
- Migrate Admin Tests - #7545		
- Add docker build and push to build job - #7547		
- Migrating tests to use Client Factory	 - #7543		
- Remove reference to Hub - #7548		
- move ListPipelineInfo to ppsutil to reduce PPS master dependency on APIServer	- #7529		
- Fix `pachctl unmount --all` bug for mac - #7306		
- Add responses for mount/unmount handlers - #7403		
- Skip validation errors when choosing a job's base commit. - #7519		
- Use prettier to autoformat the circleci config and helm values files - #7523		
- [ENT-89] Azure cluster creation script - #7512	
- Turn off mockIDP role binding when using Upstream IDPs - #7484	
- NewSQLTupleWriter and tests - #7507	
- Make pipeline logic and side-effect handling explicit	 - #7489	
- Clean up container after proto build again. - #7511	
- Batch copy operations in unordered writer. - #7436	
- Fix test after merge.	- #7505		
- Expose configuration for compaction sharding - #7503	
- Ensure shard path ranges are disjoint. - #7474		
- Start to break out composite uses of k8s in pipeline_controller into kubeDriver - #7449		
- Fix missing files (flaky correctness issue) & switch keys to mount name - #7428		
- pachctl port-forward: cleanup on SIGHUP (window closed) - #7500		
- Exit finishing early for errored commits.	 - #7433		
- pachsql: GetTableInfo	 - #7439
- don't include deployTarget in enterprise server's helm config	 - #7479	
- Add buffering to debug dump stream	 - #7493		
- Significantly improve performance in FUSE mounts	 - #7495	
- Decouple content defined chunking from file batching	 - #7442		
- Remove the restart call from inside getRC	 - #7475		
- Use a singleton pach address object in deploy.go	 - #7476	
- Support Dex Secret files for IDPs	 - #7466		
- Re-revert wikipedia example license matches - #7478	
- Expose pachd.identityDatabaseFullNameOverride to configure dex DB name - #7464	
- Wordcount example now finds license 13 times - #7465	
- PPS master event nits	 - #7446		
- expose PGBOUNCER_DEFAULT_POOL_SIZE - #7454	
- Fix typo and improve verbiage in pachctl pps cmds	- #7409	
- Address CVE-2021-3121 - #7453	
- Speed up pause/unpause test - #7451		
- Find helm chart in a more os-independent way.	 - #7445	
- Block task service setup on etcd client creation.	 - #7448		
- Refactor to reduce complexity in Pipeline Controller	 - #7438	
- Fix pachtcl port-forward so that it directs by default at the pod ports - #7443		
- Remove deleted load tests from circle	 - #7437	
- Distribute load testing - #7388	
- Structured Data: CSV & JSON Parsers - #7370		
- Start simplifying pipeline_controller.go - #7393
- storage/chunk: Fix TaskChain executing callbacks after Wait has returned.	 - #7416	
- Propagating Custom certs referenced in pachd's SSL_CERT_DIR to side-cars	 - #7414
- Add unmounting all repos handler - #7302	
- Strip build directory from pachd stack traces	 - #7413	
- [mount server] Error if trying to mount nonexistent repo - #7389	
- Don't propagate TotalFileSet if parent commit errored in alias commits  - #7408	
- Add .png extension to charts - #7396	
- Fix Secure (HTTPS) setting for Minio	- #7390		
- Upgrade to latest S2 library - #7397		
- [ENT-102] pachgen is non-deterministic	 - #7387	
- [ENT-39] Implement pause and unpause	 - #7272		
- DeleteAll() before and after test in AcquireClient - #7395	
- Pool pachyderm clusters for parallel integration tests - #7357	
- Fix: check sql.ErrNoRows in error chain - #7380
- Set Nodeports to the same values as the service port - #7366
- Receive keep alive responses for worker ip renewal - #7367	
- Fix debug dump segfault on failed jobs - #7373	
- Add post upgrade hook for configpod - #7375	
- Remove Enterprise Server's init container	 - #7368	
- [Mount Server] Add 401 return code for unauthenticated user - #7365	
- Update to Pachyderm 2.1.0, fix PostgreSQL password - #7362	
- CORE-288 Add Snowflake driver - #7346	
- pachctl: only print "reading from stdin" if stdin is a terminal - #7356	
- pps/cmds: change jsonnet args to StringArrayVar from StringSliceVar - #7336	
- Limit k8s requests inside getRC retry. - #7359	
- Add annotations to helm chart for pachd and console services.	 - #7345	
- Handle "never" cron ticks. - #7353
- Reduce the chunk batch size - #7351	
- Finer grained locking of Index Writer data - #7290	
- Testing enterprise deployments - #7299	
- Remove pachd's init container	 - #7335	
- Add node selectors to many components. - #7339	
- Temporarily skip ECS test. - #7341	
- Open commits performance changes  - #7333
- Add protoc plugin for autogenerating some protobuf boilerplate - #7244	
- Start Enterprise Heartbeat routine immediately - #7310
- Allow incorrect use of update pipeline. - #7332	
- Fix symlink upload with lazy files - #7301	
- Fix a couple task service issues - #7315	

## 2.1.9
- worker Master deletes meta files for unskippable datums. - #7648
- account for chart version in upgrade test installs. - #7649

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

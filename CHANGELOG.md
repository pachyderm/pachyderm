
# Changelog

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

### [What’s new](https://docs.pachyderm.com/2.0.x/getting_started/whats_new/)

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

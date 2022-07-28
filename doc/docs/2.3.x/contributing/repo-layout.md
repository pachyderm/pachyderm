# Repo layout

Following is a layout of the various directories that make up the pachyderm
repo, and their purpose.

## ETC
```shell
📦etc
 ┣ 📂build
 ┣ 📂compile
 ┣ 📂contributing
 ┣ 📂deploy
 ┃ ┣ 📂azure
 ┃ ┣ 📂cloudfront
 ┃ ┣ 📂gcp
 ┃ ┣ 📂gpu
 ┃ ┣ 📂tracing
 ┣ 📂examples
 ┣ 📂generate-envoy-config
 ┣ 📂helm
 ┃ ┣ 📂.cr-index
 ┃ ┣ 📂LICENSES
 ┃ ┣ 📂examples
 ┃ ┣ 📂pachyderm
 ┃ ┃ ┣ 📂charts
 ┃ ┃ ┣ 📂dependencies
 ┃ ┃ ┃ ┗ 📂postgresql
 ┃ ┃ ┃ ┃ ┣ 📂charts
 ┃ ┃ ┃ ┃ ┃ ┗ 📂common
 ┃ ┃ ┃ ┃ ┃ ┃ ┣ 📂templates
 ┃ ┃ ┃ ┃ ┃ ┃ ┃ ┣ 📂validations
 ┃ ┃ ┃ ┃ ┣ 📂ci
 ┃ ┃ ┃ ┃ ┣ 📂files
 ┃ ┃ ┃ ┃ ┃ ┣ 📂conf.d
 ┃ ┃ ┃ ┃ ┃ ┣ 📂docker-entrypoint-initdb.d
 ┃ ┃ ┃ ┃ ┣ 📂templates
 ┃ ┃ ┣ 📂templates
 ┃ ┃ ┃ ┣ 📂cloudsqlAuthProxy
 ┃ ┃ ┃ ┣ 📂console
 ┃ ┃ ┃ ┣ 📂enterprise-server
 ┃ ┃ ┃ ┣ 📂etcd
 ┃ ┃ ┃ ┣ 📂ingress
 ┃ ┃ ┃ ┣ 📂kube-event-tail
 ┃ ┃ ┃ ┣ 📂pachd
 ┃ ┃ ┃ ┃ ┣ 📂rbac
 ┃ ┃ ┃ ┣ 📂pgbouncer
 ┃ ┃ ┃ ┣ 📂proxy
 ┃ ┃ ┃ ┣ 📂tests
 ┃ ┣ 📂test
 ┣ 📂kube
 ┣ 📂kubernetes-kafka
 ┃ ┣ 📂0configure
 ┃ ┣ 📂2rbac-namespace-default
 ┃ ┣ 📂3zookeeper
 ┃ ┣ 📂4kafka
 ┃ ┣ 📂5outside-services
 ┣ 📂kubernetes-prometheus
 ┣ 📂netcat
 ┣ 📂proto
 ┃ ┣ 📂pachgen
 ┣ 📂redhat
 ┣ 📂test-images
 ┣ 📂testing
 ┃ ┣ 📂artifacts
 ┃ ┣ 📂circle
 ┃ ┣ 📂dags
 ┃ ┣ 📂images
 ┃ ┃ ┗ 📂ubuntu_with_s3_clients
 ┃ ┣ 📂introspect
 ┃ ┣ 📂kafka
 ┃ ┣ 📂loads
 ┃ ┃ ┣ 📂few-commits
 ┃ ┃ ┃ ┣ 📂few-modifications
 ┃ ┃ ┃ ┃ ┣ 📂few-directories
 ┃ ┃ ┃ ┃ ┣ 📂many-directories
 ┃ ┃ ┃ ┃ ┗ 📂one-directory
 ┃ ┃ ┃ ┗ 📂many-modifications
 ┃ ┃ ┃ ┃ ┗ 📂one-directory
 ┃ ┃ ┣ 📂many-commits
 ┃ ┃ ┃ ┣ 📂few-modifications
 ┃ ┃ ┃ ┗ 📂many-modifications
 ┃ ┣ 📂migration
 ┃ ┃ ┣ 📂v1_11
 ┃ ┃ ┣ 📂v1_7
 ┃ ┣ 📂opa-policies
 ┃ ┣ 📂s3gateway
 ┃ ┃ ┣ 📂runs
 ┃ ┣ 📂spout
 ┣ 📂worker
```

## SRC 

```shell
📦src # Source code 
 ┣ 📂admin 
 ┣ 📂auth 
 ┣ 📂client # protobufs & source code for go client
 ┃ ┣ 📂limit
 ┣ 📂debug
 ┣ 📂enterprise
 ┣ 📂identity
 ┣ 📂internal 
 ┃ ┣ 📂ancestry # package that parses git ancestry references
 ┃ ┣ 📂backoff # package that implements backoff algorithms for retrying operations
 ┃ ┣ 📂cert # library for generating x509 certificates
 ┃ ┣ 📂clientsdk # package for implementing gRPC APIs functions
 ┃ ┣ 📂clusterstate # package containing set of migrations for running pachd at the current version
 ┃ ┣ 📂cmdutil # utilities for pachctl CLI
 ┃ ┣ 📂collection # collection of utilities (errors, errorutil, tracing, & watch)
 ┃ ┣ 📂config # package for handling pachd config 
 ┃ ┣ 📂dbutil # utilities for handling database connections
 ┃ ┣ 📂deploy # package that detects if we're using a non-released version of pachd image
 ┃ ┣ 📂dlock # package that implements a distributed lock on top of etcd
 ┃ ┣ 📂dockertestenv # package for handling docker test environments 
 ┃ ┣ 📂errors # package for handling errors + stack traces
 ┃ ┃ ┣ 📂testing
 ┃ ┣ 📂errutil # utilities for handling error messages 
 ┃ ┣ 📂exec # package that runs external commands
 ┃ ┣ 📂fsutil # utilities for handling temporary files 
 ┃ ┣ 📂grpcutil # utilities for working with gRPC clients/servers
 ┃ ┣ 📂keycache # package that watches, caches, and returns keys in atomic value
 ┃ ┣ 📂lease # package that manages resources via leases
 ┃ ┣ 📂license # package that handles checking enterprise licensing 
 ┃ ┣ 📂log # package that formats logs and makes them pretty
 ┃ ┣ 📂lokiutil # utilities for leveraging loki logs 
 ┃ ┃ ┣ 📂client 
 ┃ ┣ 📂metrics # package that submits user & cluster metrics to segment
 ┃ ┣ 📂middleware 
 ┃ ┃ ┣ 📂auth
 ┃ ┃ ┣ 📂errors
 ┃ ┃ ┣ 📂logging
 ┃ ┃ ┗ 📂version
 ┃ ┣ 📂migrations #  package that handles env and state structs 
 ┃ ┣ 📂minikubetestenv  # package for handling minikube test environments 
 ┃ ┣ 📂miscutil # utilities for miscellaneous 
 ┃ ┣ 📂obj # package for handling objects (local, minio, amazon, cache, etc)
 ┃ ┃ ┣ 📂integrationtests
 ┃ ┣ 📂pacherr # package to check if error exists 
 ┃ ┣ 📂pachhash # package for handling hashes 
 ┃ ┣ 📂pachsql # package for handling sql ingest tool (snowflake, mysql,pgx)
 ┃ ┣ 📂pachtmpl # package for handling jsonnet templates 
 ┃ ┣ 📂pager # package that pages content to whichever pager is defined by the PAGER env-var
 ┃ ┣ 📂pbutil # utilities for working with protobufs
 ┃ ┣ 📂pfsdb  # package that contains the database schema that PFS uses.
 ┃ ┣ 📂pfsfile # package that converts paths to a canonical form used in the driver
 ┃ ┣ 📂pfsload # package that contains several pachyderm file system utilities 
 ┃ ┣ 📂pfssync # package that contains the standard PFS downloader interface 
 ┃ ┣ 📂pool # package that handles pool grpc connections & counts outstanding datums
 ┃ ┣ 📂ppsconsts # package that contains global constants used across Pachyderm
 ┃ ┣ 📂ppsdb  # package that contains the database schema that PPS uses 
 ┃ ┣ 📂ppsload # package for handling pipeline creation 
 ┃ ┣ 📂ppsutil # utilities for handling pipeline-related tasks
 ┃ ┣ 📂pretty # utilities for pretty printing durations, bytes, & progress bars
 ┃ ┣ 📂profileutil # utilities for exporting performance information to external systems
 ┃ ┣ 📂progress # package for handling progress bars 
 ┃ ┣ 📂promutil # utilities for collecting Prometheus metrics
 ┃ ┣ 📂random # pakage for returning a cryptographically random, URL safe string with length
 ┃ ┣ 📂randutil # utilities for handling unique/random strings (uuid)
 ┃ ┣ 📂require # utilities for making unit tests terser
 ┃ ┣ 📂sdata # package for handling Tuple, an alias for []interface{} used for passingx rows of data
 ┃ ┃ ┣ 📂csv
 ┃ ┣ 📂secrets # package for obfuscating secret data from being logged
 ┃ ┣ 📂serde # package for Pachyderm-specific data structures used to un/marshall go structs & maps
 ┃ ┣ 📂serviceenv # package for handling connections to other services in the cluster
 ┃ ┣ 📂storage # collection of packages that handle storage
 ┃ ┃ ┣ 📂chunk
 ┃ ┃ ┣ 📂fileset
 ┃ ┃ ┃ ┣ 📂index
 ┃ ┃ ┣ 📂kv
 ┃ ┃ ┣ 📂metrics
 ┃ ┃ ┣ 📂renew
 ┃ ┃ ┗ 📂track
 ┃ ┣ 📂stream # package for handling, comparing, and enqueing streams
 ┃ ┣ 📂tabwriter
 ┃ ┣ 📂tarutil # utilities for tar archiving 
 ┃ ┣ 📂task # package for handling the distributed processing of tasks.
 ┃ ┃ ┣ 📂cmds
 ┃ ┃ ┣ 📂taskprotos
 ┃ ┣ 📂testetcd # package for running end-to-end pachyderm tests entirely locally
 ┃ ┣ 📂testpachd
 ┃ ┣ 📂testsnowflake
 ┃ ┣ 📂testutil # utilities for [tbd]
 ┃ ┃ ┣ 📂local
 ┃ ┃ ┣ 📂random
 ┃ ┣ 📂tls
 ┃ ┣ 📂tracing
 ┃ ┃ ┣ 📂extended
 ┃ ┣ 📂transactiondb
 ┃ ┣ 📂transactionenv
 ┃ ┃ ┣ 📂txncontext
 ┃ ┣ 📂transforms
 ┃ ┣ 📂uuid
 ┃ ┗ 📂watch
 ┣ 📂license
 ┣ 📂pfs
 ┣ 📂pps 
 ┣ 📂proxy
 ┣ 📂server
 ┃ ┣ 📂admin
 ┃ ┃ ┣ 📂cmds
 ┃ ┃ ┗ 📂server
 ┃ ┣ 📂auth
 ┃ ┃ ┣ 📂cmds
 ┃ ┃ ┣ 📂server
 ┃ ┃ ┃ ┣ 📂testing
 ┃ ┃ ┣ 📂testing
 ┃ ┣ 📂cmd
 ┃ ┃ ┣ 📂mount-server
 ┃ ┃ ┃ ┣ 📂cmd
 ┃ ┃ ┣ 📂pachctl
 ┃ ┃ ┃ ┣ 📂cmd
 ┃ ┃ ┃ ┣ 📂shell
 ┃ ┃ ┣ 📂pachctl-doc
 ┃ ┃ ┣ 📂pachd
 ┃ ┃ ┣ 📂pachtf
 ┃ ┃ ┗ 📂worker
 ┃ ┣ 📂config
 ┃ ┣ 📂debug
 ┃ ┃ ┣ 📂cmds
 ┃ ┃ ┣ 📂server
 ┃ ┃ ┗ 📂shell
 ┃ ┣ 📂enterprise
 ┃ ┃ ┣ 📂cmds
 ┃ ┃ ┣ 📂limits
 ┃ ┃ ┣ 📂metrics
 ┃ ┃ ┣ 📂server
 ┃ ┃ ┣ 📂testing
 ┃ ┃ ┣ 📂text
 ┃ ┣ 📂identity
 ┃ ┃ ┣ 📂cmds
 ┃ ┃ ┣ 📂server
 ┃ ┣ 📂identityutil # utilities for [tbd]
 ┃ ┣ 📂license
 ┃ ┃ ┣ 📂cmds
 ┃ ┃ ┣ 📂server
 ┃ ┣ 📂pfs
 ┃ ┃ ┣ 📂cmds
 ┃ ┃ ┣ 📂fuse
 ┃ ┃ ┣ 📂pretty
 ┃ ┃ ┣ 📂s3
 ┃ ┃ ┣ 📂server
 ┃ ┃ ┃ ┣ 📂testing
 ┃ ┣ 📂pps
 ┃ ┃ ┣ 📂cmds
 ┃ ┃ ┣ 📂pretty
 ┃ ┃ ┣ 📂server
 ┃ ┣ 📂proxy
 ┃ ┃ ┗ 📂server
 ┃ ┣ 📂transaction
 ┃ ┃ ┣ 📂cmds
 ┃ ┃ ┣ 📂pretty
 ┃ ┃ ┣ 📂server
 ┃ ┃ ┃ ┣ 📂testing
 ┃ ┣ 📂worker
 ┃ ┃ ┣ 📂common
 ┃ ┃ ┣ 📂datum
 ┃ ┃ ┣ 📂driver
 ┃ ┃ ┣ 📂logs
 ┃ ┃ ┣ 📂pipeline
 ┃ ┃ ┃ ┣ 📂service
 ┃ ┃ ┃ ┣ 📂spout
 ┃ ┃ ┃ ┗ 📂transform
 ┃ ┃ ┣ 📂server
 ┃ ┃ ┣ 📂stats
 ┣ 📂task
 ┣ 📂templates
 ┣ 📂testing
 ┃ ┣ 📂deploy
 ┃ ┣ 📂loadtest
 ┃ ┃ ┣ 📂obj
 ┃ ┃ ┃ ┣ 📂build
 ┃ ┃ ┃ ┣ 📂cmd
 ┃ ┃ ┃ ┃ ┗ 📂supervisor
 ┃ ┃ ┃ ┣ 📂kube
 ┃ ┗ 📂match
 ┣ 📂transaction
 ┗ 📂version
 ┃ ┣ 📂versionpb
```

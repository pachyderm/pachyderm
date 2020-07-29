# Repo layout

Following is a layout of the various directories that make up the pachyderm
repo, and their purpose.

```
build
debian
doc - the Pachyderm documentation built with mkdocs
├── pachctl - cobra auto-generated docs on command-line usage
etc - everything else
├── build - scripts for building releases
├── compatibility - contains mappings of pachyderm versions to the dash versions they're compatible with
├── compile - scripts to facilitate compiling and building docker images
├── contributing - contains helper scripts/assets for contributors
├── deploy - scripts/assets for pachyderm deployments
│   ├── cloudfront
│   ├── gpu - scripts to help enable GPU resources on k8s/pachyderm
│   └── tracing - k8s manifests for enabling Jaeger tracing of pachyderm
├── initdev - scripts to stand up a vagrant environment for pachyderm
├── kube - internal scripts for working with k8s
├── kubernetes-kafka
├── kubernetes-prometheus
├── netcat
├── plugin
│   ├── logging
│   └── monitoring
├── proto - scripts for compiling protobufs
├── testing - scripts/assets used for testing
│   ├── artifacts - static assets used in testing/mocking
│   ├── deploy - scripts to assist in deploying pachyderm on various cloud providers
│   ├── entrypoint
│   ├── migration - sample data used in testing pachyderm migrations
│   ├── s3gateway - scripts for running conformance tests on the s3gateway
│   └── vault-s3-client
├── user-job
└── worker
examples - example projects; see readme for details of each one
src - source code
├── client - contains protobufs and the source code for pachyderm's go client
│   ├── admin - admin-related functionality
│   │   └── 1_7 - old, v1.7-compatible protobufs
│   ├── auth - auth-related functionality
│   ├── debug - debug-related functionality
│   ├── deploy - deployment-related functionality
│   ├── enterprise - enterprise-related functionality
│   ├── health - health check-related functionality
│   ├── limit - limit-related functionality
│   ├── pfs - PFS-related functionality
│   ├── pkg - utility packages
│   │   ├── config - pachyderm config file reading/writing
│   │   ├── discovery
│   │   ├── grpcutil - utilities for working with gRPC clients/servers
│   │   ├── pbutil - utilities for working with protobufs
│   │   ├── require - utilities for making unit tests terser
│   │   ├── shard
│   │   └── tracing - facilitates pachyderm cluster Jaeger tracing
│   ├── pps - PPS-related functionality
│   └── version - version check-related functionality
├── plugin
│   └── vault
│       ├── etc
│       ├── pachyderm
│       ├── pachyderm-plugin
│       └── vendor - vendored libraries for the vault plugin
├── server - contains server-side logic and CLI
│   ├── admin - cluster admin functionality
│   │   ├── cmds - cluster admin CLI
│   │   └── server - cluster admin server
│   ├── auth - auth functionality
│   │   ├── cmds - auth CLI
│   │   ├── server - auth server
│   │   └── testing - a mock auth server used for testing
│   ├── cmd - contains the various pachyderm entrypoints
│   │   ├── pachctl - the CLI entrypoint
│   │   ├── pachctl-doc - helps generate docs for the CLI
│   │   ├── pachd - the server entrypoint
│   │   └── worker - the worker entrypoint
│   ├── debug - debug functionality
│   │   ├── cmds - debug CLI
│   │   └── server - debug server
│   ├── deploy - storage secret deployment server
│   ├── enterprise - enterprise functionality
│   │   ├── cmds - enterprise CLI
│   │   └── server - enterprise server
│   ├── health - health check server
│   ├── http - PFS-over-HTTP server, used by the dash to serve PFS content
│   ├── pfs - PFS functionality
│   │   ├── cmds - PFS CLI
│   │   ├── fuse - support mounting PFS repos via FUSE
│   │   ├── pretty - pretty-printing of PFS metadata in the CLI
│   │   ├── s3 - the s3gateway, an s3-like HTTP API for serving PFS content
│   │   └── server - PFS server
│   ├── pkg - utility packages
│   │   ├── ancestry - parses git ancestry reference strings
│   │   ├── backoff - backoff algorithms for retrying operations
│   │   ├── cache - a gRPC server for serving cached content
│   │   ├── cert - functionality for generating x509 certificates
│   │   ├── cmdutil - functionality for helping creating CLIs
│   │   ├── collection - etcd collection management
│   │   ├── dag - a simple in-memory directed acyclic graph data structure
│   │   ├── deploy - functionality for deploying pachyderm
│   │   │   ├── assets - generates k8s manifests and other assets used in deployment
│   │   │   ├── cmds - deployment CLI
│   │   │   └── images - handling of docker images
│   │   ├── dlock - distributed lock on etcd
│   │   ├── errutil - utility functions for error handling
│   │   ├── exec - utilities for running external commands
│   │   ├── hashtree - a Merkle tree library
│   │   ├── lease - utility for managing resources with expirable leases
│   │   ├── localcache - a concurrency-safe local disk cache
│   │   ├── log - logging utilities
│   │   ├── metrics - cluster metrics service using segment.io
│   │   ├── migration
│   │   ├── netutil - networking utilities
│   │   ├── obj - tools for working with various object stores (e.g. S3)
│   │   ├── pfsdb - the etcd database schema that PFS uses
│   │   ├── pool - gRPC connection pooling
│   │   ├── ppsconsts - PPS-related constants
│   │   ├── ppsdb - the etcd database schema that PPS uses
│   │   ├── ppsutil - PPS-related utility functions
│   │   ├── pretty - function for pretty printing values
│   │   ├── serviceenv - management of connections to pach services
│   │   ├── sql - tools for working with postgres database dumps
│   │   ├── sync - tools for syncing PFS content
│   │   ├── tabwriter - tool for writing tab-delimited content
│   │   ├── testutil - test-related utilities
│   │   ├── uuid - UUID generation
│   │   ├── watch - tool for watching etcd databases for changes
│   │   └── workload
│   ├── pps - PPS functionality
│   │   ├── cmds - - PPS CLI
│   │   ├── example - example PPS requests
│   │   ├── pretty - pretty printing of PPS output to the CLI
│   │   └── server - PPS server
│   │       └── githook - support for github PPS sources
│   ├── vendor - vendored packages
│   └── worker - pachd master and sidecar
└── testing - testing tools
    ├── loadtest - load tests for pachyderm
    │   └── split - stress tests of PFS merge functionality
    ├── match - a grep-like tool used in testing
    ├── saml-idp
    └── vendor - vendored packages
```

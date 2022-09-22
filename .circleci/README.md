# Pachyderm – CI, Build & Release Documentation

## Patch Release Instructions

TODO

## Minor Release Instructions

TODO

## Updating Vendor Docker Mirrors/Images

### [pachyderm/etcd](https://hub.docker.com/repository/docker/pachyderm/etcd)
---
etcd uses [gcr.io/etcd-development/etcd]() as a primary container registry, and [quay.io/coreos/etcd]() as secondary. visit [etcd's github](https://github.com/etcd-io/etcd/releases) to determine the version you wish to update our mirror to.

Next trigger the circle ci workflow to upgrade it. passing the following workflow parameters.

| name | type | value |
| --| --| --|
|release-etcd | boolean | `true` |
|etcd-image-version | string | `v3.5.5` |

After verifying the workflow ran without issues, verify the new etcd image has been published to the [pachyderm/etcd](https://hub.docker.com/repository/docker/pachyderm/etcd) docker repo. You can then open a new PR to update the etcd refrences in pachyderm to run CI and merge for the next release. 

### [pachyderm/pgbouncer](https://hub.docker.com/repository/docker/pachyderm/pgbouncer)
---

Our pgbouncer image is built by embedding the pgbouncer binary, an layering in tls support. To build and publish a new version of pgbouncer, make sure the new version is availiabe at http://www.pgbouncer.org/ and trigger a new circle ci workflow with the following pipeline paramaters. 

| name | type | value |
| --| --| --|
|release-pgbouncer | boolean | `true` |
|pgbouncer-image-version | string | `1.17.0` |

After verifying the workflow ran without issues, verify the new pgbouncer image has been published to the [pachyderm/pgbouncer](https://hub.docker.com/repository/docker/pachyderm/pgbouncer) docker repo. You can then open a new PR to update the pgbouncer refrences in pachyderm to run CI and merge for the next release.

### [pachyderm/postgresql](https://hub.docker.com/repository/docker/pachyderm/postgresql)
---

TODO

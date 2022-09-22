# Pachyderm – CI, Build & Release Documentation

## Patch Release Instructions

TODO

## Minor Release Instructions

TODO

## Updating Vendor Docker Mirrors

### [pachyderm/etcd](https://hub.docker.com/repository/docker/pachyderm/etcd)
---
etcd uses [gcr.io/etcd-development/etcd]() as a primary container registry, and [quay.io/coreos/etcd]() as secondary. visit [etcd's github](https://github.com/etcd-io/etcd/releases) to determine the version you wish to update our mirror to.

Next trigger the circle ci workflow to upgrade it. passing the following workflow parameters.

| name | type | value |
| --| --| --|
|release-etcd | boolean | `true` |
|etcd-image-version | string | `v3.5.5` |

After verifying the workflow ran without issues, verify the new etcd image has been published to the [pachyderm/etcd]() docker repo. You can then open a new PR to update the etcd refrences in pachyderm to run CI and merge for the next release. 

### [pachyderm/pgbouncer](https://hub.docker.com/repository/docker/pachyderm/pgbouncer)
---

TODO

### [pachyderm/postgresql](https://hub.docker.com/repository/docker/pachyderm/postgresql)
---

TODO

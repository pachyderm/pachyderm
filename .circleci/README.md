# Pachyderm – CI, Build & Release Documentation

## Patch Release Instructions

TODO: Add detailed instructions

1. bump the console version in [values.yaml](https://github.com/pachyderm/pachyderm/blob/master/etc/helm/pachyderm/values.yaml#L72) before release
2. simply create a tag on release branch. i.e `v2.3.5`
3. approve the approval job on the triggered circle workflow 

## Minor Release Instructions

TODO

## Alpha Release Instructions

TODO

## Updating Vendor Docker Mirrors/Images

## [pachyderm/etcd](https://hub.docker.com/repository/docker/pachyderm/etcd)
---
etcd uses [gcr.io/etcd-development/etcd]() as a primary container registry, and [quay.io/coreos/etcd]() as secondary. visit [etcd's github](https://github.com/etcd-io/etcd/releases) to determine the version you wish to update our mirror to.

Next trigger the circle ci workflow to upgrade it. passing the following workflow parameters.

| name | type | value |
| --| --| --|
|release-etcd | boolean | `true` |
|etcd-image-version | string | `v3.5.5` |

After verifying the workflow ran without issues, verify the new etcd image has been published to the [pachyderm/etcd](https://hub.docker.com/repository/docker/pachyderm/etcd) docker repo. You can then open a new PR to update the etcd refrences in pachyderm to run CI and merge for the next release. 

## [pachyderm/pgbouncer](https://hub.docker.com/repository/docker/pachyderm/pgbouncer)
---

Our pgbouncer image is built by embedding the pgbouncer binary, an layering in tls support. To build and publish a new version of pgbouncer, make sure the new version is availiabe at http://www.pgbouncer.org/ and trigger a new circle ci workflow with the following pipeline paramaters. 

| name | type | value |
| --| --| --|
|release-pgbouncer | boolean | `true` |
|pgbouncer-image-version | string | `1.17.0` |

After verifying the workflow ran without issues, verify the new pgbouncer image has been published to the [pachyderm/pgbouncer](https://hub.docker.com/repository/docker/pachyderm/pgbouncer) docker repo. You can then open a new PR to update the pgbouncer refrences in pachyderm to run CI and merge for the next release.

## [pachyderm/postgresql](https://hub.docker.com/repository/docker/pachyderm/postgresql)
---

Our postgresql is a fork repo from [bitnami](https://github.com/bitnami/bitnami-docker-postgresql) which has since been archieved. NOTE: this is for the built in postgres that is bundled in, This setup is not recommended in production environments. 

Releases are maintain in the internal [pachyderm/postgresql repo](https://github.com/pachyderm/postgresql)

## [pachyderm/kube-event-tail](https://hub.docker.com/repository/docker/pachyderm/kube-event-tail)
---

Our kube-event-tail image is built from the source [kube-event-tail internal repo](https://github.com/pachyderm/kube-event-tail) with the kube-event-tail as a submodule.

In order to update this image, simply go to the [kube-event-tail internal repo](https://github.com/pachyderm/kube-event-tail) and create a tag in the format `v0.0.8`, circle ci will trigger, build and release the new image. Once it is published you can bump the image in the core repo and test before merging for the next release. 

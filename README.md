# Notice 

**This repo is currently in the process of being moved into the [Pachyderm Repo](https://github.com/pachyderm/pachyderm) Please submit pull requests and issues there.**

Currently this repo is still used for GH releases and artifacthub.

Migration Status:
- [x] Pachyderm 2.x chart (in master branch, under etc/helm)
- [ ] Pachyderm 1.x chart (will be moved to 1.13.x branch line)
# Pachyderm Helm Chart

The repo contains the pachyderm helm chart.

## Usage

Create a `values.yaml` file with your storage provider of choice, and other options, then run

```shell
$ helm repo add pach https://pachyderm.github.io/helmchart
$ helm install pachd pach/pachyderm -f values.yaml
```

## Developer Guide
To run the tests for the helm chart, run the following:

```shell
$ make test
```

# Diff against `pachctl`

To see how this Helm chart and `pachctl deploy` differ, one can do
something similar to the following:

1. Generate a `pachctl` manifest with

    ```shell
    $ pachctl deploy googleBUCKET-NAME 10 --dynamic-etcd-nodes 1 -o yaml --dry-run > pachmanifest.yaml
    ```

1. Generate a Helm manifest with

    ```shell
    $ helm template -f examples/gcp-values.yaml ./pachyderm > helmmanifest.yaml
    ```

1. Visually diff the two.

# JSON Schema

We use [helm-schema-gen](https://github.com/karuppiah7890/helm-schema-gen)
to manage the JSON schema.  It can be installed with:

```shell
$ helm plugin install https://github.com/karuppiah7890/helm-schema-gen.git
```

When updating `values.yaml` please run the following to update the
json schema file.

```shell
$ cd pachyderm
$ helm schema-gen values.yaml > values.schema.json
```

# Validate Helm manifest

```shell
go install github.com/instrumenta/kubeval
kubeval helmmanifest.yaml
```

<!-- SPDX-FileCopyrightText: Pachyderm, Inc. <info@pachyderm.com>
SPDX-License-Identifier: Apache-2.0 -->

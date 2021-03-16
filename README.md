# Pachyderm Helm Chart

The repo contains the pachyderm helm chart.

Status: **Experimental**

## Folder Structure

```
pachyderm - the helm chart itself
test - Go based tests for the helm chart
```

## Developer Guide
To run the tests for the helm chart, run the following:

```
make test
```

# Diff against `pachctl`

To see how this Helm chart and `pachctl deploy` differ, one can do
something similar to the following:

 1. Generate a `pachctl` manifest with `pachctl deploy google
    BUCKET-NAME 10 --dynamic-etcd-nodes 1 -o yaml --dry-run >
    pachmanifest.yaml`
 2. Generate a Helm manifest with `helm template -f
    examples/gcp-values.yaml ./pachyderm > helmmanifest.yaml`
 3. Visually diff the two.

# JSON Schema

We use [helm-schema-gen](https://github.com/karuppiah7890/helm-schema-gen)
to manage the JSON schema.  It can be installed with:

```sh
$ helm plugin install https://github.com/karuppiah7890/helm-schema-gen.git
```

When updating `values.yaml` please run the following to update the
json schema file.

```sh
cd pachyderm
helm schema-gen values.yaml > values.schema.json
```

# Validate Helm manifest
 1. `go install github.com/instrumenta/kubeval`
 2. `kubeval helmmanifest.yaml`

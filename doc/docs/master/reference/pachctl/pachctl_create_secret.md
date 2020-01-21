## pachctl create secret

Create a kubernetes secret

### Synopsis

Create a kubernetes secret when one does not have direct kubectl access to the kubernetes cluster. The newly created secret can then be referenced by the `transform.image_pull_secrets` field on a pipeline spec in order to pull docker images from a private registry whose credentials are conained in the secret. For details on using `iamge_pull_secrets`, see https://docs.pachyderm.com/latest/reference/pipeline_spec/#transform-required.

```
pachctl create secret [flags]
```

### Options

```
  -f, --file string        The JSON or YAML file containing the kubernetes secret, it can be a url or local file. - reads from stdin. (default "-")
  -n, --namespace          The Kubernetes namespace of secret (default is "default" namespace)
```

### Options inherited from parent commands

```
      --no-color   Turn off colors.
  -v, --verbose    Output verbose logs
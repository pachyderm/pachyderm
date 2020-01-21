## pachctl create secret

Create a kubernetes secret.

### Synopsis

Creates a Kubernetes secret when direct `kubectl` access to the Kubernetes cluster is not available. Then, the newly created secret can be referenced by the `transform.image_pull_secrets` field in a pipeline spec to pull docker images from a private registry whose credentials are contained in the secret. For 
 more information about `image_pull_secrets`, see https://docs.pachyderm.com/latest/reference/pipeline_spec/#transform-required.

```
pachctl create secret [flags]
```

### Options

```
  -f, --file string        The JSON or YAML file that contains the Kubernetes secret, it can be a URL or local file. - reads from stdin. (default "-")
  -n, --namespace          The Kubernetes namespace of secret (default is "default" namespace)
```

### Options inherited from parent commands

```
      --no-color   Turn off colors.
  -v, --verbose    Output verbose logs.

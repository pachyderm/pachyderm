## pachctl config import-kube

Import a kubernetes context as a Pachyderm context, and set the active Pachyderm context.

### Synopsis

Import a kubernetes context as a Pachyderm context. By default the current kubernetes context is used.

```
pachctl config import-kube <context> [flags]
```

### Options

```
  -e, --enterprise          Configure an enterprise server context.
  -h, --help                help for import-kube
  -k, --kubernetes string   Specify the kubernetes context's values to import.
  -n, --namespace string    Specify a namespace where Pachyderm is deployed.
  -o, --overwrite           Overwrite a context if it already exists.
```

### Options inherited from parent commands

```
      --no-color   Turn off colors.
  -v, --verbose    Output verbose logs
```


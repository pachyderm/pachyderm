## pachctl config set context

Set a context.

### Synopsis

Set a context config from a given name and either JSON stdin, or a given kubernetes context.

```
pachctl config set context <context> [flags]
```

### Options

```
  -h, --help                help for context
  -k, --kubernetes string   Import a given kubernetes context's values into the Pachyderm context.
      --overwrite           Overwrite a context if it already exists.
```

### Options inherited from parent commands

```
      --no-color   Turn off colors.
  -v, --verbose    Output verbose logs
```


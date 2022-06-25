## pachctl mount-server

Start a mount server for controlling FUSE mounts via a local REST API.

### Synopsis

Starts a REST mount server, running in the foreground and logging to stdout.

```
pachctl mount-server [flags]
```

### Options

```
  -h, --help               help for mount-server
      --mount-dir string   Target directory for mounts e.g /pfs (default "/pfs")
```

### Options inherited from parent commands

```
      --no-color   Turn off colors.
  -v, --verbose    Output verbose logs
```


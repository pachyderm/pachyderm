## pachctl undeploy

Tear down a deployed Pachyderm cluster.

### Synopsis

Tear down a deployed Pachyderm cluster.

```
pachctl undeploy [flags]
```

### Options

```
  -a, --all                
                           Delete everything, including the persistent volumes where metadata
                           is stored.  If your persistent volumes were dynamically provisioned (i.e. if
                           you used the "--dynamic-etcd-nodes" flag), the underlying volumes will be
                           removed, making metadata such repos, commits, pipelines, and jobs
                           unrecoverable. If your persistent volume was manually provisioned (i.e. if
                           you used the "--static-etcd-volume" flag), the underlying volume will not be
                           removed.
  -h, --help               help for undeploy
      --namespace string   Kubernetes namespace to undeploy Pachyderm from.
```

### Options inherited from parent commands

```
      --no-color   Turn off colors.
  -v, --verbose    Output verbose logs
```


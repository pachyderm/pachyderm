## pachctl auth set

Set the scope of access that 'username' has to 'repo'

### Synopsis

Set the scope of access that 'username' has to 'repo'. For example, 'pachctl auth set github-alice none private-data' prevents "github-alice" from interacting with the "private-data" repo in any way (the default). Similarly, 'pachctl auth set github-alice reader private-data' would let "github-alice" read from "private-data" but not create commits (writer) or modify the repo's access permissions (owner). Currently all Pachyderm authentication uses GitHub OAuth, so 'username' must be a GitHub username

```
pachctl auth set <username> (none|reader|writer|owner) <repo> [flags]
```

### Options

```
  -h, --help   help for set
```

### Options inherited from parent commands

```
      --no-color   Turn off colors.
  -v, --verbose    Output verbose logs
```


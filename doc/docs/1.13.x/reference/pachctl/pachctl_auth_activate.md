## pachctl auth activate

Activate Pachyderm's auth system

### Synopsis

Activate Pachyderm's auth system, and restrict access to existing data to the
user running the command (or the argument to --initial-admin), who will be the
first cluster admin

```
pachctl auth activate [flags]
```

### Options

```
  -h, --help                   help for activate
      --initial-admin string   The subject (robot user or github user) who
                               will be the first cluster admin; the user running 'activate' will identify as
                               this user once auth is active.  If you set 'initial-admin' to a robot
                               user, pachctl will print that robot user's Pachyderm token; this token is
                               effectively a root token, and if it's lost you will be locked out of your
                               cluster
```

### Options inherited from parent commands

```
      --no-color   Turn off colors.
  -v, --verbose    Output verbose logs
```


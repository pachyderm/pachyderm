This folder contains bash scripts for deploying a test pachyderm cluster on a
cloud provider. All scripts must conform to the following API:

```
./aaa.sh --create      # Creates a new test cluster in the cloud provider aaa.
                       # In this cluster, pachyderm is up and running, and
                       # kubectl is connected to the cluster. When run with this
                       # flag, aaa.sh must print a cluster ID on a single line
                       # by itself after it finishes (unless there is an error)
                       # This will let us capture the cluster ID with
                       # ID=$(./aaa --create | tail -n1)

./aaa.sh --delete=<id> # Deletes a test cluster identified by <id>. This must be
                       # the same ID printed at the end of --create. This will
                       # allow a user to run
                       # ID=$( ./aaa.sh --create | tail -n1 ); ./aaa.sh --delete=$ID

./aaa.sh --delete-all  # Deletes any existing test clusters in the cloud provider
                       # aaa
```

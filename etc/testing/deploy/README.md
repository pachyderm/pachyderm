This folder contains bash scripts for deploying a test pachyderm cluster on a
cloud provider. All scripts must conform to the following API:

```
./aaa.sh --delete-all # Deletes any existing test clusters in the cloud provider
                      # aaa

./aaa.sh --create     # Creates a new test cluster in the cloud provider aaa.
                      # In this cluster, pachyderm is up and running, and
                      # kubectl is connected to the cluster.
```

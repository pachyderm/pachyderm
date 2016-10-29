## Backup

Please backup your metadata storage system before running this script.  See the [migration guide](http://pachyderm.readthedocs.io/en/latest/production/migration.html) for details.

## How to run this script

Make sure that your `kubectl` has been configured correctly to talk to the Kubernetes cluster where your Pachyderm runs.  Then simply:

```
make run
```

You should see a pod named `migrate` being created:

```
kubectl get all
```

Once the pod finishes running, the migration is complete.  You can make sure that nothing went wrong by looking at the logs of the pod:

```
kubectl logs migrate
```

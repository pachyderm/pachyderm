# Troubleshooting

## Triaging

### Check portforwarding

If you see the following:

```
$ kubectl get all
The connection to the server localhost:8080 was refused - did you specify the right host or port?
```

You probably have not enabled port forwarding. You can start port forwarding by running something like:

```
ssh <HOST> -fTNL 8080:localhost:8080 -L 30650:localhost:30650
```

Where `<HOST>` is one of the names of your docker hosts. You can see a list by running:

```
docker-machine ls
```

### Check your versions

```
$ pachctl version
COMPONENT           VERSION             
pachctl             1.1.0               
pachd               1.1.0 
```

This tells us:

* the pachctl client is installed
* the pachyderm cluster is up and serving traffic
* it also tells us which versions of each we're running (they don't necessarily need to match)

If this looks normal, [continue triaging](#check-pachyderm-cluster-status)

If the command errors:

```
$ pachctl version
-bash: /Users/sjezewski/go/bin/pachctl: No such file or directory
```

You don't have `pachctl` installed. [See the Installation Troubleshooting Section](#installation)

If you get a response like:

```
$ pachctl version
COMPONENT           VERSION             
pachctl             1.1.0               
pachd               (version unknown) : error connecting to pachd server at address (0.0.0.0:30650): context deadline exceeded

please make sure pachd is up (`kubectl get all`) and portforwarding is enabled
```

You aren't connected to the pachyderm server. [Refer to the Deployment Troubleshooting Section](#not-connected-to-pachd)

### Check Pachyderm Cluster Status

If you want to validate that the Pachyderm cluster is up and ready, you can run:

```
$ ./etc/kube/check_pachd_ready.sh
All pods are ready.
```

If you see something like:

```
$ ./etc/kube/check_pachd_ready.sh
Empty result
```

Then you probably haven't deployed the Pachyderm cluster. [Follow the deployment instructions](./deploying.html)

### Check k8s Cluster Status

If you have deployed, but you don't see the 'ready' message above, check the health of the cluster:

```
kubectl get all
```

A healthy cluster looks something like this:

```
NAME                                     DESIRED      CURRENT            AGE
etcd                                     1            1                  16m
pachd                                    2            2                  16m
rethink                                  1            1                  16m
NAME                                     CLUSTER-IP   EXTERNAL-IP        PORT(S)                        AGE
etcd                                     10.0.0.5     <none>             2379/TCP,2380/TCP              16m
kubernetes                               10.0.0.1     <none>             443/TCP                        66d
pachd                                    10.0.0.46    nodes              650/TCP                        16m
rethink                                  10.0.0.191   nodes              8080/TCP,28015/TCP,29015/TCP   16m
NAME                                     READY        STATUS             RESTARTS                       AGE
etcd-wwynx                               1/1          Running            0                              16m
k8s-etcd-127.0.0.1                       1/1          Running            0                              66d
k8s-master-127.0.0.1                     4/4          Running            0                              66d
k8s-proxy-127.0.0.1                      1/1          Running            0                              66d
pachd-loe59                              1/1          Running            1                              16m
pachd-u3nse                              1/1          Running            0                              16m
rethink-qk1km                            1/1          Running            0                              16m
```

The things to look at for debugging are the two pachd pods (`pachd-u3nse` and `pachd-loe59`) and the rethink pod (`rethink-qk1km`).

All of these should have status `Running`. If any of them do not, look at the logs, e.g:

```
$ kubectl logs pachd-loe59
```

This should give you some indication about why its failing to startup. 


## Installation

First make sure you followed the [installation steps](./installation.html)

Then make sure that your installation location is on your path.

Your installation path is:

- `usr/local/bin` if you installed via homebrew
- `usr/bin` if you installed via deb package
- `$GOPATH/bin` if you installed from source

To check your path:
```
$ echo $PATH
/usr/local/opt/coreutils/libexec/gnubin:/usr/local/go/bin:/usr/local/bin:/usr/bin:/bin:/usr/sbin:/sbin:/usr/local/bin:/opt/X11/bin:/usr/local/go/bin:/Users/myusername/go/bin
```

Make sure that your installation path is listed as one of the locations in `$PATH`. If it's not, update your `PATH` variable in your shell profile, for example:

```
export PATH=/usr/local/go/bin:$PATH
```




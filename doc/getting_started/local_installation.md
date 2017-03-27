# Local Installation
This guide will walk you through the recommended path to get Pachyderm running locally on OSX or Linux.

If you hit any errors not covered in this guide, check our [troubleshooting](http://pachyderm.readthedocs.io/en/stable/getting_started/troubleshooting.html) docs for common errors, submit an issue on [GitHub](https://github.com/pachyderm/pachyderm), join our users channel on Slack, or email us at [support@pachyderm.io](mailto:support@pachyderm.io) and we can help you right away.

## Prerequisites
- [Minikube](#minikube) (and VirtualBox)
- [Pachyderm Command Line Interface](#pachctl)

### Minikube

Kubernetes offers a fantastic guide to [install minikube](http://kubernetes.io/docs/getting-started-guides/minikube). Follow the Kubernetes installation guide to install Virtual Box, Minikibe, and Kubectl. Then come back here to install Pachyderm.

### Pachctl

`pachctl` is a command-line utility used for interacting with a Pachyderm cluster.


```shell
# For OSX:
$ brew tap pachyderm/tap && brew install pachyderm/tap/pachctl@1.3

# For Linux (64 bit):
$ curl -o /tmp/pachctl.deb -L https://github.com/pachyderm/pachyderm/releases/download/v1.3.17/pachctl_1.3.17_amd64.deb && sudo dpkg -i /tmp/pachctl.deb
```

To check that installation was successful, you can try running `pachctl help`, which should return a list of Pachyderm commands.

## Deploy Pachyderm
Now that you have Minikube running, it's incredibly easy to deploy Pachyderm.

```sh
pachctl deploy local
```
This generates a Pachyderm manifest and deploys Pachyderm on Kubernetes. It may take a few minutes for the pachd nodes to be running because it's pulling containers from DockerHub. You can see the cluster status by using `kubectl get all`:

```sh
$ kubectl get all
NAME               READY     STATUS    RESTARTS   AGE
po/etcd-xzc0d      1/1       Running   0          55s
po/pachd-6m6wm     1/1       Running   0          55s
po/rethink-388b3   1/1       Running   0          55s

NAME         DESIRED   CURRENT   READY     AGE
rc/etcd      1         1         1         55s
rc/pachd     1         1         1         55s
rc/rethink   1         1         1         55s

NAME             CLUSTER-IP   EXTERNAL-IP   PORT(S)                                          AGE
svc/etcd         10.0.0.92    <none>        2379/TCP,2380/TCP                                55s
svc/kubernetes   10.0.0.1     <none>        443/TCP                                          9m
svc/pachd        10.0.0.61    <nodes>       650:30650/TCP,651:30651/TCP                      55s
svc/rethink      10.0.0.87    <nodes>       8080:32080/TCP,28015:32081/TCP,29015:32085/TCP   55s

NAME              DESIRED   SUCCESSFUL   AGE
jobs/pachd-init   1         1            55s
```
Note: If you see a few restarts on the pachd nodes, that's ok. That simply means that Kubernetes tried to bring up those containers before Rethink was ready so it restarted them.

### Port Forwarding

The last step is to set up port forwarding so commands you send can reach Pachyderm within the VM. We background this process since port forwarding blocks.

```shell
$ pachctl port-forward &
```

Once port forwarding is complete, pachctl should automatically be connected. Try `pachctl version` to make sure everything is working.

```shell
$ pachctl version
COMPONENT           VERSION
pachctl             1.3.2
pachd               1.3.2
```

We're good to go!


## Next Steps

Now that you have everything installed and working, check out our [Beginner Tutorial](./beginner_tutorial.html) to learn the basics of Pachyderm such as adding data and building analysis pipelines.




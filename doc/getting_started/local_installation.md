# Local Installation
This guide will walk you through the recommended path to get Pachyderm running locally on OSX or Linux. 

If you hit any errors not covered in this guide, check our :doc:`troubleshooting` docs for common errors, submit an issue on [GitHub](https://github.com/pachyderm/pachyderm), join our users channel on Slack, or email us at [support@pachyderm.io](mailto:support@pachyderm.io) and we can help you right away.  

## Prerequisites
- [Minikube](#minikube) (and VirtualBox)
- [Pachyderm Command Line Interface](#pachctl)

### Minikube

Kubernetes offers a fantastic guide to [install minikube](http://kubernetes.io/docs/getting-started-guides/minikube). Follow the Kubernetes installation guide to install Virtual Box, Minikibe, and Kubectl. Then come back here to install Pachyderm. 

### Pachctl

`pachctl` is a command-line utility used for interacting with a Pachyderm cluster.


```shell
# For OSX:
$ brew tap pachyderm/tap && brew install pachctl

# For Linux (64 bit):
$ curl -o /tmp/pachctl.deb -L https://pachyderm.io/pachctl.deb && dpkg -i /tmp/pachctl.deb
```

To check that installation was successful, you can try running `pachctl help`, which should return a list of Pachyderm commands.

## Deploy Pachyderm
Now that you have Minikube running, it's incredibly easy to deploy Pachyderm.

```sh
pachctl deploy
```
This generates a Pachyderm manifest and deploys Pachyderm on Kubernetes. It may take a few minutes for the pachd nodes to be running because it's pulling containers from DockerHub. You can see the cluster status by using `kubectl get all`:

```sh
$ kubectl get all
NAME            DESIRED      CURRENT       AGE
etcd            1            1             6s
pachd           2            2             6s
rethink         1            1             6s
NAME            CLUSTER-IP   EXTERNAL-IP   PORT(S)                        AGE
etcd            10.0.0.45    <none>        2379/TCP,2380/TCP              6s
kubernetes      10.0.0.1     <none>        443/TCP                        6m
pachd           10.0.0.101   <nodes>       650/TCP                        6s
rethink         10.0.0.182   <nodes>       8080/TCP,28015/TCP,29015/TCP   6s
NAME            READY        STATUS        RESTARTS                       AGE
etcd-swoag      1/1          Running       0                              6s
pachd-7xyse     1/1          Running       0                              6s
pachd-gfdc6     1/1          Running       0                              6s
rethink-v5rsx   1/1          Running       0                              6s
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
pachctl             1.2.0
pachd               1.2.0
```

We're good to go!


## Next Steps

Now that you have everything installed and working, check out our [Beginner Tutorial](./beginner_tutorial.html) to learn the basics of Pachyderm such as adding data and building analysis pipelines. 




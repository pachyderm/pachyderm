# Amazon Web Services

### Prerequisites

- [AWS CLI](https://aws.amazon.com/cli/) - have it installed and have your [AWS credentials](http://docs.aws.amazon.com/cli/latest/userguide/cli-chap-getting-started.html) configured.
- [kubectl](https://kubernetes.io/docs/user-guide/prereqs/)
- [kops](https://github.com/kubernetes/kops/blob/master/docs/install.md)
- [jq](https://stedolan.github.io/jq/download/)
- uuid

### Install `pachctl`

To deploy and interact with Pachyderm, you will need `pachctl`, a command-line utility used for Pachyderm. To install `pachctl` run one of the following:


```shell
# For OSX:
$ brew tap pachyderm/tap && brew install pachyderm/tap/pachctl@1.4

# For Linux (64 bit):
$ curl -o /tmp/pachctl.deb -L https://github.com/pachyderm/pachyderm/releases/download/v1.4.0/pachctl_1.4.0_amd64.deb && sudo dpkg -i /tmp/pachctl.deb
```

You can try running `pachctl version` to check that this worked correctly, but Pachyderm itself isn't deployed yet so you won't get a `pachd` version.

```sh
$ pachctl version
COMPONENT           VERSION
pachctl             1.4.0
pachd               (version unknown) : error connecting to pachd server at address (0.0.0.0:30650): context deadline exceeded.
```

### Deploy Pachyderm

The easiest way to deploy a Pachyderm cluster on AWS is with our deploy script. Once you have the prerequisites mentioned above, dowload and run our AWS deploy script by running:

```
curl -o aws.sh https://github.com/pachyderm/pachyderm/blob/master/etc/deploy/aws.sh
chmod +x aws.sh
sudo -E ./aws.sh
```

This script will use kops to deploy Kubernetes and Pachyderm in AWS.  The script will ask you for your AWS credentials, region preference, etc.  If you would like to customize the number of nodes in the cluster, node types, etc., you can open up the deploy script and modify the respective fields.

The script will take a few minutes, and Pachyderm will take an addition couple of minutes to spin up.  Once it is up, `kubectl get all` should return something like:

```
NAME             READY     STATUS    RESTARTS   AGE
po/etcd-wn317    1/1       Running   0          5m
po/pachd-mljp6   1/1       Running   3          5m

NAME       DESIRED   CURRENT   READY     AGE
rc/etcd    1         1         1         5m
rc/pachd   1         1         1         5m

NAME             CLUSTER-IP   EXTERNAL-IP   PORT(S)                         AGE
svc/etcd         10.0.0.165   <nodes>       2379:32379/TCP,2380:32686/TCP   5m
svc/kubernetes   10.0.0.1     <none>        443/TCP                         5m
svc/pachd        10.0.0.214   <nodes>       650:30650/TCP,651:30651/TCP     5m
```

Finally, we need to set up forward a port so that pachctl can talk to the cluster.

```sh
# Forward the ports. We background this process because it blocks.
$ pachctl port-forward &
```

And you're done! You can test to make sure the cluster is working by trying `pachctl version`:
```sh
$ pachctl version
COMPONENT           VERSION
pachctl             1.4.0
pachd               1.4.0
```


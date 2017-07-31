# Deploying Enterprise Edition

The primary interface for Enterprise Edition is the Pachyderm Dashboard.  Pachyderm Enterprise deployment is as simple as deploying the dashboard on top of or along with a Pachyderm cluster:

- [Deploying on top of an existing Pachyderm deployment](#deploying-on-top-of-an-existing-pachyderm-deployment)
- [Deploying with a new Pachyderm deployment](#deploying-with-a-new-pachyderm-deployment)

## Deploying on top of an existing Pachyderm deployment

Assuming you followed one of our [deploy guides](http://pachyderm.readthedocs.io/en/latest/deployment/deploy_intro.html) and you have a Pachyderm cluster running, you should see that the state of your Pachyderm cluster is similar to the following:

```
$ kubectl get all
NAME                        READY     STATUS    RESTARTS   AGE
po/etcd-4197107720-br61m    1/1       Running   0          8m
po/pachd-3548222380-s086m   1/1       Running   2          8m

NAME             CLUSTER-IP     EXTERNAL-IP   PORT(S)                       AGE
svc/etcd         10.111.11.36   <nodes>       2379:32379/TCP                8m
svc/kubernetes   10.96.0.1      <none>        443/TCP                       10m
svc/pachd        10.97.116.5    <nodes>       650:30650/TCP,651:30651/TCP   8m

NAME           DESIRED   CURRENT   UP-TO-DATE   AVAILABLE   AGE
deploy/etcd    1         1         1            1           8m
deploy/pachd   1         1         1            1           8m

NAME                  DESIRED   CURRENT   READY     AGE
rs/etcd-4197107720    1         1         1         8m
rs/pachd-3548222380   1         1         1         8m
```

Your `pachctl` CLI tool should also be able to connect to Pachyderm:

```
$ pachctl version
COMPONENT           VERSION             
pachctl             1.5.0               
pachd               1.5.0
```

Once you have Pachyderm in this state, deploying Pachyderm Enterprise Edition is as simple as:

```
$ pachctl deploy local --dashboard-only
```

**Note** - Even though you might not have deployed Pachyderm locally, you can run `pachctl deploy local --dashboard-only` to deploy the Pachyderm Dashboard on any Pachyderm cluster (as long as `pachctl` is connect to that cluster).  This includes clusters deployed on AWS, Google Cloud, Azure, or on premise. 

After a few minutes, you should see the new `dash-xxxxxxxx` pod running in Kubernetes:

```
$ kubectl get all
NAME                        READY     STATUS    RESTARTS   AGE
po/dash-3809689541-56tfb    2/2       Running   0          8m
po/etcd-4197107720-jd9w5    1/1       Running   0          13m
po/pachd-4280389576-qfcg8   1/1       Running   3          13m

NAME             CLUSTER-IP   EXTERNAL-IP   PORT(S)                         AGE
svc/dash         10.0.0.4     <nodes>       8080:30080/TCP,8081:30081/TCP   8m
svc/etcd         10.0.0.152   <nodes>       2379:32379/TCP                  13m
svc/kubernetes   10.0.0.1     <none>        443/TCP                         17m
svc/pachd        10.0.0.193   <nodes>       650:30650/TCP,651:30651/TCP     13m

NAME           DESIRED   CURRENT   UP-TO-DATE   AVAILABLE   AGE
deploy/dash    1         1         1            1           8m
deploy/etcd    1         1         1            1           13m
deploy/pachd   1         1         1            1           13m

NAME                  DESIRED   CURRENT   READY     AGE
rs/dash-3809689541    1         1         1         8m
rs/etcd-4197107720    1         1         1         13m
rs/pachd-4280389576   1         1         1         13m
```

If you previously had port forwarding enabled for your Pachyderm cluster, you will also have to restart this forwarding:

```
$ ps aux | grep "forward"
dwhitena 12625  0.0  0.3 475896 26204 pts/13   SNl  14:51   0:00 pachctl port-forward
dwhitena 12674  0.0  0.4  58232 35208 pts/13   SNl  14:51   0:00 kubectl port-forward pachd-4280389576-qfcg8 30650:650
dwhitena 15467  0.0  0.0  14224  1040 pts/13   S+   15:06   0:00 grep --color=auto --exclude-dir=.bzr --exclude-dir=CVS --exclude-dir=.git --exclude-dir=.hg --exclude-dir=.svn forward
$ kill 12674
Terminated                                                                                                                                                                                                              
UI not enabled, deploy with --dashboard
[1]  + 12625 exit 1     pachctl port-forward
$ pachctl port-forward &
[1] 15481
Pachd port forwarded
Dash websocket port forwarded
Dash UI port forwarded, navigate to localhost:38080
CTRL-C to exit
```

Now you can visit the Pachyderm dashboard at `localhost:38080`!  The dashboard will prompt you for your enterprise token that you received when registering for Pachyderm Enterprise Edition:

![alt tag](token.png)

## Deploying with a new Pachyderm deployment

You can deploy Pachyderm Enterprise Edition with any new Pachyderm deployment by adding the `--dashboard` flag to the respective deploy command:

```
# AWS
pachctl deploy amazon ... --dashboard

# Google
pachctl deploy google ... --dashboard

# Azure
pachctl deploy azure ... --dashboard

# Local
pachctl deploy local --dashboard

# Custom
pachctl deploy custom ... --dashboard
```

Each of these deploys are further detailed [here](http://pachyderm.readthedocs.io/en/latest/deployment/deploy_intro.html).  

After deploying with those commands, you should see the `dash-xxxxxxxxx` pod running in Kubernetes, and you should be able to access the dashboard at `localhost:38080`, as discussed further [above](#deploying-on-top-of-an-existing-pachyderm-deployment).

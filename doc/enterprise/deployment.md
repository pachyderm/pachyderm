# Deploying Enterprise Edition

To use Pachyderm's Enterprise Edition features, you simply need to [activate them](#activating-pachyderm-enterprise-edition) on a deployed pachyderm cluster.  Also, assuming you want to use the Pachyderm Enterprise Dashboard features (which is the primary interface to Pachyderm Enterprise), you will need to [deploy the dashboard](#deploying-the-pachyderm-enterprise-edition-dashboard) on top of or along with your Pachyderm cluster.

## Activating Pachyderm Enterprise Edition

There are two ways to activate Pachyderm's enterprise features::

- [Activate Pachyderm Enterprise via the `pachctl` CLI](#activate-via-the-pachctl-cli)
- [Activate Pachyderm Enterprise via the dashboard](#activate-via-the-dashboard)

For either method, you will need to have your Pachyderm Enterprise activation code available.  You should have received this from Pachyderm sales/support when registering for the Enterprise Edition.  Please contact [support@pachyderm.io](mailto:support@pachyderm.io) if you are having trouble locating your activation code. 

### Activate via the `pachctl` CLI

Assuming you followed one of our [deploy guides](http://pachyderm.readthedocs.io/en/latest/deployment/deploy_intro.html) and you have a Pachyderm cluster running, you should see that the state of your Pachyderm cluster is similar to the following:

```
$ kubectl get all
NAME                       READY     STATUS    RESTARTS   AGE
po/dash-361776027-vbj73    2/2       Running   0          1h
po/etcd-2142892294-whlpn   1/1       Running   0          1h
po/pachd-776177201-ktjlv   1/1       Running   0          1h

NAME             CLUSTER-IP   EXTERNAL-IP   PORT(S)                                     AGE
svc/dash         10.0.0.91    <nodes>       8080:30080/TCP,8081:30081/TCP               1h
svc/etcd         10.0.0.231   <nodes>       2379:32379/TCP                              1h
svc/kubernetes   10.0.0.1     <none>        443/TCP                                     2h
svc/pachd        10.0.0.136   <nodes>       650:30650/TCP,651:30651/TCP,652:30652/TCP   1h

NAME           DESIRED   CURRENT   UP-TO-DATE   AVAILABLE   AGE
deploy/dash    1         1         1            1           1h
deploy/etcd    1         1         1            1           1h
deploy/pachd   1         1         1            1           1h

NAME                 DESIRED   CURRENT   READY     AGE
rs/dash-361776027    1         1         1         1h
rs/etcd-2142892294   1         1         1         1h
rs/pachd-776177201   1         1         1         1h 
```

You should also be able to connect to the Pachyderm cluster via the `pachctl` CLI:

```
$ pachctl version
COMPONENT           VERSION             
pachctl             1.6.0           
pachd               1.6.0
```

Activating the Enterprise features of Pachyderm is then as easy as:

```
$ pachctl enterprise activate <activation-code>
```

If this command returns no error, then the activation was successful. The state of the Enterprise activation can also be retrieved at any time via:

```
$ pachctl enterprise get-state   
ACTIVE
```  

### Activate via the dashboard

Assuming that you have a running Pachyderm cluster and you have deployed the Pachyderm Enterprise dashboard using [this guide](#deploying-the-pachyderm-enterprise-edition-dashboard), you should be able to visit `<pachyderm host IP>:30080` (e.g., `localhost:30080` when you are using `pachctl port-forward`) to see the dashboard. When you first visit the dashboard, it will prompt you for your activation code:

![alt tag](token.png)

Once you enter your activation code, you should have full access to the Enterprise dashboard and your cluster will be an active Enterprise Edition cluster.  This could be confirmed with:

```
$ pachctl enterprise get-state   
ACTIVE
```

## Deploying the Pachyderm Enterprise Edition Dashboard

The Pachyderm Enterprise dashboard can be deployed on top of an existing Pachyderm cluster or along with the deployment of a new cluster:

- [Deploying the dashboard on top of an existing Pachyderm deployment](#deploying-on-top-of-an-existing-pachyderm-deployment)
- [Deploying the dashboard with a new Pachyderm deployment](#deploying-with-a-new-pachyderm-deployment)

### Deploying on top of an existing Pachyderm deployment

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
pachctl             1.6.0               
pachd               1.6.0
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
$ pachctl port-forward &
```

Now you can visit the Pachyderm dashboard at `localhost:30080`! 

### Deploying with a new Pachyderm deployment

You can deploy the Pachyderm Enterprise Edition dashboard with any new Pachyderm deployment by adding the `--dashboard` flag to the respective deploy command:

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

After deploying with those commands, you should see the `dash-xxxxxxxxx` pod running in Kubernetes, and you should be able to access the dashboard at `localhost:30080`, as discussed further [above](#deploying-on-top-of-an-existing-pachyderm-deployment).

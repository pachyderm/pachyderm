# Deploying Enterprise Edition

To deploy and use Pachyderm's Enterprise Edition, you simply need to follow [one of our guides to deploy Pachyderm](../deployment/overview.html) and then [activate the Enterprise Edition](#activating-pachyderm-enterprise-edition).

**Note** - Pachyderm's Enterprise dashboard is now deployed by default with Pachyderm. If you wish to deploy without the dashboard please use `pachctl deploy [command] --no-dashboard`

**Note** - You can get a FREE evaluation token for the enterprise edition on the landing page of the Enterprise dashboard.

## Activating Pachyderm Enterprise Edition

There are two ways to activate Pachyderm's enterprise features::

- [Activate Pachyderm Enterprise via the `pachctl` CLI](#activate-via-the-pachctl-cli)
- [Activate Pachyderm Enterprise via the dashboard](#activate-via-the-dashboard)

For either method, you will need to have your Pachyderm Enterprise activation code available.  You should have received this from Pachyderm sales/support when registering for the Enterprise Edition.  If you are a new user evaluating Pachyderm, you can receive a FREE evaluation code on the landing page of the dashboard. Please contact [support@pachyderm.io](mailto:support@pachyderm.io) if you are having trouble locating your activation code. 

### Activate via the `pachctl` CLI

Assuming you followed one of our [deploy guides](http://pachyderm.readthedocs.io/en/latest/deployment/deploy_intro.html) and you have a Pachyderm cluster running, you should see that the state of your Pachyderm cluster is similar to the following:

```
$ kubectl get pods
NAME                     READY     STATUS    RESTARTS   AGE
dash-6c9dc97d9c-vb972    2/2       Running   0          6m
etcd-7dbb489f44-9v5jj    1/1       Running   0          6m
pachd-6c878bbc4c-f2h2c   1/1       Running   0          6m
```

You should also be able to connect to the Pachyderm cluster via the `pachctl` CLI:

```
$ pachctl version
COMPONENT           VERSION             
pachctl             1.6.8           
pachd               1.6.8
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


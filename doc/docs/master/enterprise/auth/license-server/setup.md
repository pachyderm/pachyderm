# Enterprise Server Setup
The **Enterprise Server** is a component in Pachyderm which manages Enterprise Licensing
and the integration with a company's Identity Providers (IDPs).

At a high level, an organization can have **many Pachyderm clusters federated under one single Enterprise Server**. Administrators activate the Enterprise Server with an **Enterprise License Key** from Pachyderm sales, and optionally configure authentication with their IDP via SAML, OIDC, LDAP, etc...

The following diagram gives you a quick overview of an organization with multiple Pachyderm clusters behind a single Enterprise Server.
![Enterprise Server General Deployment](../images/enterprise-server.png)

For POCs and smaller organizations with one single Pachyderm cluster, the **Enterprise Server services can be run embedded in pachd**. A separate deployment is not necessary. An organization with a single Pachyderm cluster can run the Enterprise Server services embedded within pachd.

The setup of an Enterprise Server requires to:

1. Deploy it.
1. Activate your Enterprise Key and enable Auth.
1. Register your newly created or existing Pachyderm clusters with your enterprise server.
1. Optional: Enable Auth on each cluster.

## 1 - Deploy An Enterprise Server

### Single-cluster deployment
Deploying a Pachyderm cluster with the embedded enterprise server does not require any specific action.
Proceed as usual:

1. [Install your favorite version of `pachctl`](https://docs.pachyderm.com/latest/getting_started/local_installation/#install-pachctl).
1. [Deploy Pachyderm](https://docs.pachyderm.com/latest/getting_started/local_installation/#deploy-pachyderm): `pachctl deploy <local, google.....>`.
1. [Activate your enterprise Key](https://docs.pachyderm.com/latest/enterprise/deployment/#activate-pachyderm-enterprise-edition): `pachctl enterprise activate`
1. [Enable authentication](https://docs.pachyderm.com/latest/enterprise/auth/enable-auth/#activate-access-controls-with-pachctl): `pachctl auth activate` //TODO update this link once auth2.0 is out


This results in a single pachd pod, with authentication enabled. Proceed to [configuring IDP integrations]()//TODO update this link once auth2.0 is out.

### Multi-cluster deployment

Deploying a stand-alone enterprise server requires using the `--enterprise-server` flag for `pachctl deploy`. 

- If a pachyderm cluster will also be installed in the same kubernetes cluster, they should be installed in **different namespaces**:

	```shell
	$ kubectl create namespace enterprise
	$ pachctl deploy <local, google...> --enterprise-server --namespace enterprise
	```

	This command deploys postgres, etcd and a deployment and service called `pach-enterprise`. 
	`pach-enterprise` uses the same docker image and pachd binary, but it **listens on a different set of ports (31650, 31657, 31658)** to avoid conflicts with pachd.

- Check the state of your deployment by running:
	```shell
	kubectl get all --namespace enterprise
	```
	**System Response**
	```
	NAME                                   READY   STATUS    RESTARTS   AGE
	pod/etcd-5fd7c675b6-46kz7              1/1     Running   0          113m
	pod/pach-enterprise-6dc9cb8f66-rs44t   1/1     Running   0          105m
	pod/postgres-6bfd7bfc47-9mz28          1/1     Running   0          113m

	```

## 2- Activate enterprise licensing and enable auth

- Use your enterprise key to activate your enterprise server.
- Then enable Authentication at the Enterprise Server level:
	```shell
	$ pachctl auth activate --enterprise
	```

	!!! Warning
		Enabling Auth will return a `root token` for the enterprise server. 
		**This is separate from the root tokens for each pachd (cluster)**. 
		They should all be stored securely.

Once the enterprise server is deployed, 
deploy your cluster(s) (`pachctl deploy <local, google...>`) and register it/them with the enterprise server.

## 3-  Register your clusters
Run this command for each of the clusters you wish to register:

```shell
$ pachctl enterprise register --id <my-pachd> --enterprise-server-address pach-enterprise.enterprise:650 --pachd-address <pachd.default>:650
```

* `--enterprise-server-address` is the host and port where pachd can reach the enterprise server. 
In this example this is inside the kubernetes cluster. In production the enterprise server may be exposed on the internet.

* `--pachd-address` is the host and port where the enterprise server can reach pachd. 
This may be internal to the kubernetes cluster, or over the internet.

## 4- Enable Auth on each cluster
Finally, activate auth in the pachd (cluster). 
This is an optional step as clusters can be registered with the enterprise server without auth being enabled. 

```shell
$ pachctl auth activate --client-id my-pachd
```
 
To manage you server, its context, or connect your IdP, visit the [**Manage your Enterprise Server**](./manage.md) page.
Contexts

## Undeploy
- Undeploy the Enterprise Server
- Undeploy a Specific cluster
	- Unregister a cluster
	```shell
	$ pachctl license delete <cluster id>
	```
//TODO when more info - For now
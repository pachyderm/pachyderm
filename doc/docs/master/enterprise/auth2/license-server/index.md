# Enterprise Server
The **Enterprise Server** manages Enterprise Licensing and federates the integration with customer's Identity Providers (IDPs).

At a high level, an organization can have many Pachyderm clusters federated under one single Enterprise Server. Administrators activate the Enterprise Server with an Enterprise License Key from Pachyderm sales, and optionally configure authentication with their IDP via SAML, OIDC, LDAP, etc...


An organization with multiple Pachyderm clusters and a single Enterprise Server


For POCs and smaller organizations with a single Pachyderm cluster, the **Enterprise Server services can be run embedded in pachd**. A separate deployment is not necessary.

An organization with a single Pachyderm cluster can run the Enterprise Server services embedded within pachd.


## Deploying
### Single-cluster deployment
Deploying a Pachyderm cluster with the embedded enterprise server continues to work like in 1.0:

pachctl deploy local
pachctl enterprise activate 
pachctl auth activate


This results in a single pachd pod, with authentication enabled. Proceed to configuring IDP integrations.
Multi-cluster deployment

Deploying a stand-alone enterprise server requires using the `--enterprise` flag for pachctl deploy. If a pachyderm cluster will also be installed in the same kubernetes cluster, they should be installed in different namespaces:

pachctl deploy local --enterprise --namespace enterprise

This command deploys postgres, etcd and a deployment and service called pach-enterprise. Pach-enterprise uses the same docker image and pachd binary, but it listens on a different set of ports (31650, 31657, 31658) to avoid conflicts with pachd.

Configure auth and enterprise licensing for the enterprise server:


pachctl license activate
	pachctl auth activate --enterprise


Activating auth will return a root token for the enterprise server. This is separate from the root tokens for each pachd. They should all be stored securely.

Once the enterprise server is deployed, deploy pachd and register it with the enterprise server:

pachctl deploy local
pachctl enterprise register --id my-pachd --enterprise-server-address pach-enterprise.enterprise:650 --pachd-address pachd.default:650

The `--enterprise-server-address` is the host and port where pachd can reach the enterprise server. In this example this is inside the kubernetes cluster. In production the enterprise server may be exposed on the internet.

Likewise, `--pachd-address` is the host and port where the enterprise server can reach pachd. This may be internal to the kubernetes cluster, or over the internet.

Finally, activate auth in the pachd. This isn’t necessary - clusters can be registered with the enterprise server without auth being enabled: 

pachctl auth activate --client-id my-pachd

Contexts
The enterprise server is a separate context in the pachctl config file. Pachctl has an active pachd context, and separately an active enterprise context. To see the active enterprise context:

pachctl config get active-enterprise-context

In a single-cluster deployment, the active enterprise context will be the same as the enterprise context.

The `pachctl  license` and `pachctl identity` commands run against the enterprise context. 
`pachctl auth` commands accept an `--enterprise` flag to run against the enterprise context.
Configuring IDPs
To configure IDP integrations, use `pachctl idp create-connector`. For example, for Github:

pachctl idp create-connector --id github --name github --type github --config - <<EOF
	{
 		 "clientID": “<id>”,
  		 "clientSecret": "<secret>",
 		 "redirectURI": "http://localhost:30658/callback"
	}
EOF

For a list of connectors and the configuration options, see: https://dexidp.io/docs/connectors/
Configuring Enterprise Server Admins
By default only the robot token can administer the Enterprise Server. To add more users as administrators, run:

	pachctl auth set enterprise clusterAdmin user:<email>
Listing registered clusters
To list all registered pachds, run:


pachctl license list-clusters
	
The output includes the pachd version, whether auth is enabled, and the last heartbeat:

id: localhost
address: localhost:650
version: 2.0.0-3fc220792931fec53866ca1620e37db600f91ba9
auth_enabled: false
last_heartbeat: 2021-03-24 15:42:14.231065 +0000 UTC
Updating the enterprise license
To apply a new license and have it picked up by all pachds, run:

	pachctl license activate

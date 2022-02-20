# Manage Your Enterprise Server

## Contexts
The enterprise server **has a separate context** in the pachctl config file (`~/.pachyderm/config.json`).

Pachctl has an active pachd context (the cluster it is binded to), 
and separately an **active enterprise context**. 

To check the active enterprise context, run:
```shell
pachctl config get active-enterprise-context
```

!!! Warning "Important Notes"
    - In a single-cluster deployment, the **active enterprise context will be the same as the enterprise context**.
    - The `pachctl  license` and `pachctl idp` commands **run against the enterprise context**. 
`pachctl auth` commands accept an `--enterprise` flag to run against the enterprise context.

## Configuring IDPs
To configure IDP integrations, use `pachctl idp create-connector` as documented in 
the [**Pachyderm Integration with Identity Providers**](../../authentication/idp-dex) page.

## Manage your Enterprise Server

### Add Users As Administrators
By default, only the `root token` (Root User) can administer the Enterprise Server. 
Run the following command to add more ClusterAdmin to your Enterprise Server:
```shell
pachctl auth set enterprise clusterAdmin user:<email>
```

### List All Registered Clusters
```shell
pachctl license list-clusters
```	
The output includes the pachd version, whether auth is enabled, and the last heartbeat:
```
id: pach-2
address: 34.71.247.191:650
version: 2.0.0
auth_enabled: true
last_heartbeat: 2021-05-21 18:43:42.157027 +0000 UTC
```

### Synchronize all available contexts in your `~/.pachyderm/config.json` file
In the case where the enterprise server of your organization has multiple pachd instances,
you can use the following command to “discover” other pachd instances. It will automatically update your `~/.pachyderm/config.json` file with all the contexts you can connect to.

```shell
pachctl enterprise sync-contexts
```	

### Update The Enterprise License
To apply a new license and have it picked up by all clusters, run:
```shell
pachctl license activate --no-register
```

### Unregister A Cluster
To unregister a given cluster from your Enterprise Server, run:
```shell
pachctl license delete-cluster --id <cluster id>
```

### Undeploy

- To undeploy a Cluster registered with an Enterprise Server: 
    - Unregister the cluster as mentioned above (`pachctl license delete-cluster`)
    - Then, undeploy it: `helm uninstall`

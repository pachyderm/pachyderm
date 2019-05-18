# OpenShift

[OpenShift](https://www.openshift.com/) is a popular enterprise Kubernetes distribution.  
Pachyderm can run on OpenShift with a few small tweaks in the deployment process, which will be outlined below.
Please see [known issues](#known-issues) below for currently issues with OpenShift deployments.

## Prerequisites

Pachyderm needs a few things to install and run successfully in any Kubernetes environment

1. A persistent volume, used by Pachyderm's `etcd` for storage of system metatada. 
   The kind of PV you provision will be dependent on your infrastructure. 
   For example, many on-premises deployments use Network File System (NFS) access to some kind of enterprise storage.
1. An object store, used by Pachyderm's `pachd` for storing all your data. 
   The object store you use will probably be dependent on where you're going to run OpenShift: S3 for [AWS](https://pachyderm.readthedocs.io/en/latest/deployment/amazon_web_services.html), GCS for [Google Cloud Platform](https://pachyderm.readthedocs.io/en/latest/deployment/google_cloud_platform.html), Azure Blob Storage for  [Azure](https://pachyderm.readthedocs.io/en/latest/deployment/azure.html), or a storage provider like Minio, EMC's ECS or Swift providing S3-compatible access to enterprise storage for on-premises deployment.
1. Access to particular TCP/IP ports for communication.

### Persistent volume

You'll need to create a persistent volume with enough space for the metadata associated with the data you plan to store Pachyderm. 
The `pachctl deploy` command for AWS, GCP and Azure creates persistent storage for you, when you follow the instructions below.
A custom deploy can also create storage.  
We'll show you below how to take out the PV that's automatically created, in case you want to create it outside of the Pachyderm deployment and just consume it.

We're currently developing good rules of thumb for scaling this storage as your Pachyderm deployment grows,
but it looks like 10G of disk space is sufficient for most purposes.

### Object store

Size your object store generously, once you start using Pachyderm, you'll start versioning all your data.
You'll need four items to configure object storage

1. The access endpoint, an url.  For example, Minio's endpoints are usually something like http://minio-server:9000/
1. The bucket name you're dedicating to Pachyderm. Pachyderm will need exclusive access to this bucket.
1. The access key id for the object store.  This is like a user name for logging into the object store.
1. The secret key for the object store.  This is like the above user's password.

### TCP/IP ports

OpenShift runs containers and pods as unprivileged users which don't have access to port numbers below 1024.
Pachyderm's default manifests use ports below 1024, so you'll have to modify the manifests to use other port numbers.
It's usually as easy as adding a "1" in front of the port numbers we use.

## Preparing to deploy Pachyderm

Things you'll need

1. Your PV.  It can be created separately.
1. Your object store information
1. Your project in OpenShift
1. A text editor for editing your deployment manifest

## Deploying Pachyderm

### 1. Setting up PV and object stores

How you deploy Pachyderm on OpenShift is largely going to depend on where OpenShift is deployed. 
Below you'll find links to the documentation for each kind of deployment you can do.
Follow the instructions there for setting up persistent volumes and object storage resources.
Don't yet deploy your manifest, come back here after you've set up your PV and object store.

* OpenShift Deployed on [AWS](https://pachyderm.readthedocs.io/en/latest/deployment/amazon_web_services.html) 
* OpenShift Deployed on [GCP](https://pachyderm.readthedocs.io/en/latest/deployment/google_cloud_platform.html)
* OpenShift Deployed on [Azure](https://pachyderm.readthedocs.io/en/latest/deployment/azure.html)
* OpenShift Deployed [on-premise](https://pachyderm.readthedocs.io/en/latest/deployment/on_premises.html)

### 2. Run the deploy command with --dry-run

Once you have your PV, object store, and project, you can create a manifest for editing using the `--dry-run` argument to `pachctl deploy`.
That step is detailed in the deployment instructions for each type of deployment, above.

Below, find an example is using AWS elastic block storage as a persistent disk with a custom deploy.
We'll show how to remove this PV in case you want to use a PV you create separately.


```
pachctl deploy custom --persistent-disk aws --object-store s3 <pv-storage-name> <pv-storage-size> <s3-bucket-name> <s3-access-key-id> <s3-access-secret-key> <s3-access-endpoint-url> --static-etcd-volume=<pv-storage-name> > manifest.json
```

### 3. Modify pachd Service ports

In the deployment manifest, which we called `manifest.json`, above, find the stanza for the `pachd` Service.  An example is shown below.
```
{
	"kind": "Service",
	"apiVersion": "v1",
	"metadata": {
		"name": "pachd",
		"namespace": "default",
		"creationTimestamp": null,
		"labels": {
			"app": "pachd",
			"suite": "pachyderm"
		},
		"annotations": {
			"prometheus.io/port": "9091",
			"prometheus.io/scrape": "true"
		}
	},
	"spec": {
		"ports": [
			{
				"name": "s3gateway-port",
				"port": 600,
				"targetPort": 0,
				"nodePort": 30600
			},
			{
				"name": "api-grpc-port",
				"port": 650,
				"targetPort": 0,
				"nodePort": 30650
			},
			{
				"name": "trace-port",
				"port": 651,
				"targetPort": 0,
				"nodePort": 30651
			},
			{
				"name": "api-http-port",
				"port": 652,
				"targetPort": 0,
				"nodePort": 30652
			},
			{
				"name": "saml-port",
				"port": 654,
				"targetPort": 0,
				"nodePort": 30654
			},
			{
				"name": "api-git-port",
				"port": 999,
				"targetPort": 0,
				"nodePort": 30999
			}
		],
		"selector": {
			"app": "pachd"
		},
		"type": "NodePort"
	},
	"status": {
		"loadBalancer": {}
	}
}
```
While the nodePort declarations are fine, the port declarations are too low for OpenShift. Good example values are shown below.
```
	"spec": {
		"ports": [
			{
				"name": "s3gateway-port",
				"port": 1600,
				"targetPort": 0,
				"nodePort": 30600
			},
			{
				"name": "api-grpc-port",
				"port": 1650,
				"targetPort": 0,
				"nodePort": 30650
			},
			{
				"name": "trace-port",
				"port": 1651,
				"targetPort": 0,
				"nodePort": 30651
			},
			{
				"name": "api-http-port",
				"port": 1652,
				"targetPort": 0,
				"nodePort": 30652
			},
			{
				"name": "saml-port",
				"port": 1654,
				"targetPort": 0,
				"nodePort": 30654
			},
			{
				"name": "api-git-port",
				"port": 1999,
				"targetPort": 0,
				"nodePort": 30999
			}
		],
```

### 4. Modify pachd Deployment ports and add environment variables
In this case you're editing two parts of the `pachd` Deployment json.  
Here, we'll omit the example of the unmodified version.
Instead, we'll show you the modified version.
#### 4.1 pachd Deployment ports
The `pachd` Deployment also has a set of port numbers in the spec for the `pachd` container. 
Those must be modified to match the port numbers you set above for each port.
```
{
	"kind": "Deployment",
	"apiVersion": "apps/v1beta1",
	"metadata": {
		"name": "pachd",
		"namespace": "default",
		"creationTimestamp": null,
		"labels": {
			"app": "pachd",
			"suite": "pachyderm"
		}
	},
	"spec": {
		"replicas": 1,
		"selector": {
			"matchLabels": {
				"app": "pachd",
				"suite": "pachyderm"
			}
		},
		"template": {
			"metadata": {
				"name": "pachd",
				"namespace": "default",
				"creationTimestamp": null,
				"labels": {
					"app": "pachd",
					"suite": "pachyderm"
				},
				"annotations": {
					"iam.amazonaws.com/role": ""
				}
			},
			"spec": {
				"volumes": [
					{
						"name": "pach-disk"
					},
					{
						"name": "pachyderm-storage-secret",
						"secret": {
							"secretName": "pachyderm-storage-secret"
						}
					}
				],
				"containers": [
					{
						"name": "pachd",
						"image": "pachyderm/pachd:1.9.0rc1",
						"ports": [
							{
								"name": "api-grpc-port",
								"containerPort": 1650,
								"protocol": "TCP"
							},
							{
								"name": "trace-port",
								"containerPort": 1651
							},
							{
								"name": "api-http-port",
								"containerPort": 1652,
								"protocol": "TCP"
							},
							{
								"name": "peer-port",
								"containerPort": 1653,
								"protocol": "TCP"
							},
							{
								"name": "api-git-port",
								"containerPort": 1999,
								"protocol": "TCP"
							},
							{
								"name": "saml-port",
								"containerPort": 1654,
								"protocol": "TCP"
							}
						],
                        
```
#### 4.2 Add environment variables
There are six environment variables necessary for OpenShift
1. `WORKER_USES_ROOT`: This controls whether worker pipelines run as the root user or not. You'll need to set it to `false`
1. `PORT`: This is the grpc port used by pachd for communication with `pachctl` and the api.  It should be set to the same value you set for `api-grpc-port` above.
1. `PPROF_PORT`: This is used for Prometheus. It should be set to the same value as `trace-port` above.
1. `HTTP_PORT`: The port for the api proxy.  It should be set to `api-http-port` above.
1. `PEER_PORT`: Used to coordinate `pachd`'s. Same as `peer-port` above.
1. `PPS_WORKER_GRPC_PORT`: Used to talk to pipelines. Should be set to a value above 1024.  The example value of 1680 below is recommended.

The added values below are shown inserted above the `PACH_ROOT` value, which is typically the first value in this array.
The rest of the stanza is omitted for clarity.
```
						"env": [
                            {
                            "name": "WORKER_USES_ROOT",
                            "value": "false"
                            },
                            {
                            "name": "PORT",
                            "value": "1650"
                            },
                            {
                            "name": "PPROF_PORT",
                            "value": "1651"
                            },
                            {
                            "name": "HTTP_PORT",
                            "value": "1652"
                            },
                            {
                            "name": "PEER_PORT",
                            "value": "1653"
                            },
                            {
                            "name": "PPS_WORKER_GRPC_PORT",
                            "value": "1680"
                            },
							{
								"name": "PACH_ROOT",
								"value": "/pach"
							},

```
### 5. Change ClusterRoles to Roles

You'll find two stanzas, `ClusterRole` and `ClusterRoleBinding`.  Default Openshift security policies require you to change those to `Role` and `RoleBinding`, respectively. You can safely do a global replace of `ClusterRole` with `Role` in your text editor; there should be 3 occurrences.

### 6. Optional: remove the PV created during the deploy command
If you're using a PV you've created separately, remove the PV that was added to your manifest by `pachctl deploy --dry-run`.  Here's the example PV we created with the deploy command we used above, so you can recognize it.
```
{
	"kind": "PersistentVolume",
	"apiVersion": "v1",
	"metadata": {
		"name": "etcd-volume",
		"namespace": "default",
		"creationTimestamp": null,
		"labels": {
			"app": "etcd",
			"suite": "pachyderm"
		}
	},
	"spec": {
		"capacity": {
			"storage": "10Gi"
		},
		"awsElasticBlockStore": {
			"volumeID": "pach-disk",
			"fsType": "ext4"
		},
		"accessModes": [
			"ReadWriteOnce"
		],
		"persistentVolumeReclaimPolicy": "Retain"
	},
	"status": {}
}
```

## 7. Deploy the Pachyderm manifest you modified.

```sh
$ oc create -f pachyderm.json
```

You can see the cluster status by using `oc get pods` as in upstream Kubernetes:

```sh
    $ oc get pods
    NAME                     READY     STATUS    RESTARTS   AGE
    dash-6c9dc97d9c-89dv9    2/2       Running   0          1m
    etcd-0                   1/1       Running   0          4m
    pachd-65fd68d6d4-8vjq7   1/1       Running   0          4m
```

### Known issues

Problems related to OpenShift deployment are tracked in [this issue](https://github.com/pachyderm/pachyderm/issues/336). 






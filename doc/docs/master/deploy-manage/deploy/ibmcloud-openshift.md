# IBM Cloud OpenShift 3.1

[OpenShift](https://www.openshift.com/products/openshift-ibm-cloud/) is a popular enterprise Kubernetes distribution.
Pachyderm can run on IBM Cloud OpenShift with a few small tweaks in the [OpenShift deployment process](./openshift.md), the most important being the StatefulSets deployment.

Please see [known issues](#known-issues) below for currently issues with OpenShift deployments.

## Prerequisites

Pachyderm needs a few things to install and run successfully in IBM Cloud OpenShift environment.

##### Binaries for CLI

- [oc](https://cloud.ibm.com/docs/openshift?topic=openshift-openshift-cli#cli_oc)
- [pachctl](../../../getting_started/local_installation/#install-pachctl)

1. Since this is a stateful set based deployment, it uses either Persistent Volume Provisioning or pre-provisioned PV's using a defined storage class.
1. An object store, used by Pachyderm's `pachd` for storing all your data.
   The object store you use will probably be dependent on where you're going to run OpenShift: S3 for [AWS](https://pachyderm.readthedocs.io/en/latest/deployment/amazon_web_services.html), GCS for [Google Cloud Platform](https://pachyderm.readthedocs.io/en/latest/deployment/google_cloud_platform.html), Azure Blob Storage for  [Azure](https://pachyderm.readthedocs.io/en/latest/deployment/azure.html), or a storage provider like Minio, EMC's ECS or Swift providing S3-compatible access to enterprise storage for on-premises deployment.
1. Access to particular TCP/IP ports for communication.

### Stateful Set

You will need a storage class as defined in IBM Cloud OpenShift cluster. You would need to specify the number of replica pods needed in the configuration.
A custom deploy can also create storage. 

!!! Warning
    The metadata service (Persistent disk) generally requires a small persistent volume size (i.e. 10GB) but **high IOPS (1500)**, therefore, depending on your disk choice, you may need to oversize the volume significantly to ensure enough IOPS.

### Object store

Size your object store generously, since this is where Pachyderm, versions and stores all your data.
You'll need four items to configure object storage, in this case we have used Minio.

1. The access endpoint.
   For example, Minio's endpoints are usually something like `minio-server:9000`. 
   Don't begin it with the protocol; it's an endpoint, not an url.
1. The bucket name you're dedicating to Pachyderm. Pachyderm will need exclusive access to this bucket.
1. The access key id for the object store.  This is like a user name for logging into the object store.
1. The secret key for the object store.  This is like the above user's password.

### TCP/IP ports

For more details on how Kubernetes networking and service definitions work, see the [Kubernetes services 
documentation](https://kubernetes.io/docs/concepts/services-networking/service/).

#### Incoming ports (port)

These are the ports internal to the containers. 
You'll find these on both the pachd and dash containers.
OpenShift runs containers and pods as unprivileged users which don't have access to port numbers below 1024.
Pachyderm's default manifests use ports below 1024, so you'll have to modify the manifests to use other port numbers.
It's usually as easy as adding a "1" in front of the port numbers we use.

#### Pod ports (targetPort)

This is the port exposed by the pod to Kubernetes, which is forwarded to the `port`.
You should leave the `targetPort` set at `0` so it will match the `port` definition. 

#### External ports (nodePorts)

This is the port accessible from outside of Kubernetes.
You probably don't need to change `nodePort` values unless your network security requirements or architecture requires you to change to another method of access. 
Please see the [Kubernetes services documentation](https://kubernetes.io/docs/concepts/services-networking/service/) for details.

## The OCPify script

A bash script that automates many of the substitutions below is available [at this gist](https://gist.github.com/gabrielgrant/86c1a5b590ae3f4b3fd32d7e9d622dc8). 
You can use it to modify a manifest created using the `--dry-run` flag to `pachctl deploy custom`, as detailed below, and then use this guide to ensure the modifications it makes are relevant to your OpenShift environment.
It requires certain prerequisites, just as [jq](https://github.com/stedolan/jq) and [sponge, found in moreutils](https://joeyh.name/code/moreutils/).

This script may be useful as a basis for automating redeploys of Pachyderm as needed. 

### Best practices: Infrastructure as code

We highly encourage you to apply the best practices used in developing software to managing the deployment process.

1. Create scripts that automate as much of your processes as possible and keep them under version control.
1. Keep copies of all artifacts, such as manifests, produced by those scripts and keep those under version control.
1. Document your practices in the code and outside it.

## Preparing to deploy Pachyderm

Things you'll need:

1. Your storage class

1. Your object store information.

1. Your project in OpenShift.

1. A text editor for editing your deployment manifest.
## Deploying Pachyderm
### 1. Determine your role security policy
Pachyderm is deployed by default with cluster roles.
Many institutional Openshift security policies require namespace-local roles rather than cluster roles.
If your security policies require namespace-local roles, use the [`pachctl deploy` command below with the `--local-roles` flag](#namespace-local-roles).
### 2. Run the deploy command with --dry-run
Once you have your PV, object store, and project, you can create a manifest for editing using the `--dry-run` argument to `pachctl deploy`.
That step is detailed in the deployment instructions for each type of deployment, above.

Below, find examples, 
with cluster roles and with namespace-local roles,
using AWS elastic block storage as a persistent disk with a custom deploy.
We'll show how to remove this PV in case you want to use a PV you create separately.

#### Cluster roles

```
pachctl deploy custom --persistent-disk aws --etcd-storage-class <storage-class-name> \
     --object-store <object-storage-name> etcd-volume <volume-size-in-gb> \
     <s3-bucket-name> <s3-access-key-id> <s3-access-secret-key> <s3-access-endpoint-url> \
     --dynamic-etcd-nodes <no-of-replica-pods> --dry-run > manifest-statefulset.json
```

#### Namespace-local roles

```
pachctl deploy custom --persistent-disk aws --etcd-storage-class <storage-class-name> \
     --object-store <object-storage-name> etcd-volume <volume-size-in-gb> \
     <s3-bucket-name> <s3-access-key-id> <s3-access-secret-key> <s3-access-endpoint-url> \
     --dynamic-etcd-nodes <no-of-replica-pods> --local-roles --dry-run > manifest-statefulset.json
```

### 4. Modify pachd Service ports

In the deployment manifest, which we called `manifest.json`, above, find the stanza for the `pachd` Service.  An example is shown below.

```
{
	"kind": "Service",
	"apiVersion": "v1",
	"metadata": {
		"name": "pachd",
		"namespace": "pachyderm",
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
			},
			{
				"name": "s3gateway-port",
				"port": 600,
				"targetPort": 0,
				"nodePort": 30600
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
			},
			{
				"name": "s3gateway-port",
				"port": 1600,
				"targetPort": 0,
				"nodePort": 30600
			}
		],
```

### 5. Modify pachd Deployment ports and add environment variables
In this case you're editing two parts of the `pachd` Deployment json.  
Here, we'll omit the example of the unmodified version.
Instead, we'll show you the modified version.
#### 5.1 pachd Deployment ports
The `pachd` Deployment also has a set of port numbers in the spec for the `pachd` container. 
Those must be modified to match the port numbers you set above for each port.

```
{
	"kind": "Deployment",
	"apiVersion": "apps/v1beta1",
	"metadata": {
		"name": "pachd",
		"namespace": "pachyderm",
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
				"namespace": "pachyderm",
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
						"image": "pachyderm/pachd:{{ config.pach_latest_version }}",
						"ports": [
							{
								"name": "api-grpc-port",
								"containerPort": 650,
								"protocol": "TCP"
							},
							{
								"name": "trace-port",
								"containerPort": 651
							},
							{
								"name": "api-http-port",
								"containerPort": 652,
								"protocol": "TCP"
							},
							{
								"name": "peer-port",
								"containerPort": 653,
								"protocol": "TCP"
							},
							{
								"name": "api-git-port",
								"containerPort": 999,
								"protocol": "TCP"
							},
							{
								"name": "saml-port",
								"containerPort": 654,
								"protocol": "TCP"
							}
						],

```

#### 5.2 Add environment variables

You need to configure the following environment variables for
OpenShift:

1. `WORKER_USES_ROOT`: This controls whether worker pipelines run as the root user or not. You'll need to set it to `false`
1. `PORT`: This is the grpc port used by pachd for communication with `pachctl` and the api.  It should be set to the same value you set for `api-grpc-port` above.
1. `HTTP_PORT`: The port for the api proxy.  It should be set to `api-http-port` above.
1. `PEER_PORT`: Used to coordinate `pachd`'s. Same as `peer-port` above.
1. `PPS_WORKER_GRPC_PORT`: Used to talk to pipelines. Should be set to a value above 1024.  The example value of 1680 below is recommended.
1. `PACH_ROOT`: The Pachyderm root directory. In an OpenShift
deployment, you need to set this value to a directory to which non-root users
have write access, which might depend on your container image. By
default, `PACH_ROOT` is set to `/pach`, which requires root privileges.
Because in OpenShift, you do not have root access, you need to modify
the default setting.

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
								"value": "<path-to-non-root-dir>"
							},

```

### 6. Statefulset configuration
It uses storage class to create a PVC based on the volume claim template in the IBM Cloud OpenShift cluster. For reference, this example is based on 'ibmc-file-silver' storage class

```
{
	"apiVersion": "apps/v1beta1",
	"kind": "StatefulSet",
	"metadata": {
		"labels": {
			"app": "etcd",
			"suite": "pachyderm"
		},
		"name": "etcd",
		"namespace": "pachyderm"
	},
	"spec": {
		"replicas": 2,
		"selector": {
			"matchLabels": {
				"app": "etcd",
				"suite": "pachyderm"
			}
		},
		"serviceName": "etcd-headless",
		"template": {
			"metadata": {
				"labels": {
					"app": "etcd",
					"suite": "pachyderm"
				},
				"name": "etcd",
				"namespace": "pachyderm"
			},
			"spec": {
				"containers": [
					{
						"args": [
							"\"/usr/local/bin/etcd\" \"--listen-client-urls=http://0.0.0.0:2379\" \"--advertise-client-urls=http://0.0.0.0:2379\" \"--data-dir=/var/data/etcd\" \"--auto-compaction-retention=1\" \"--max-txn-ops=10000\" \"--max-request-bytes=52428800\" \"--quota-backend-bytes=8589934592\" \"--listen-peer-urls=http://0.0.0.0:2380\" \"--initial-cluster-token=pach-cluster\" \"--initial-advertise-peer-urls=http://${ETCD_NAME}.etcd-headless.${NAMESPACE}.svc.cluster.local:2380\" \"--initial-cluster=etcd-0=http://etcd-0.etcd-headless.${NAMESPACE}.svc.cluster.local:2380,etcd-1=http://etcd-1.etcd-headless.${NAMESPACE}.svc.cluster.local:2380\""
						],
						"command": [
							"/bin/sh",
							"-c"
						],
						"env": [
							{
								"name": "ETCD_NAME",
								"valueFrom": {
									"fieldRef": {
										"apiVersion": "v1",
										"fieldPath": "metadata.name"
									}
								}
							},
							{
								"name": "NAMESPACE",
								"valueFrom": {
									"fieldRef": {
										"apiVersion": "v1",
										"fieldPath": "metadata.namespace"
									}
								}
							}
						],
						"image": "quay.io/coreos/etcd:v3.3.5",
						"imagePullPolicy": "IfNotPresent",
						"name": "etcd",
						"ports": [
							{
								"containerPort": 2379,
								"name": "client-port"
							},
							{
								"containerPort": 2380,
								"name": "peer-port"
							}
						],
						"resources": {
							"requests": {
								"cpu": "1",
								"memory": "2G"
							}
						},
						"volumeMounts": [
							{
								"mountPath": "/var/data/etcd",
								"name": "etcd-storage"
							}
						]
					}
				],
				"imagePullSecrets": null
			}
		},
		"volumeClaimTemplates": [
			{
				"metadata": {
					"annotations": {
						"volume.beta.kubernetes.io/storage-class": "ibmc-file-silver"
					},
					"labels": {
						"app": "etcd",
						"suite": "pachyderm"
					},
					"name": "etcd-storage",
					"namespace": "pachyderm"
				},
				"spec": {
					"accessModes": [
						"ReadWriteOnce"
					],
					"resources": {
						"requests": {
							"storage": "10Gi"
						}
					}
				}
			}
		]
	}
}
```

## 7. Deploy the Pachyderm manifest you modified.

```shell
oc create -f manifest-statefulset.json
```

You can see the cluster status by using `oc get pods` as in upstream OpenShift:

```shell
    oc get pods
    NAME                     READY     STATUS    RESTARTS   AGE
    dash-78c4b487dc-sm56p    2/2       Running   0          1m
    etcd-0                   1/1       Running   0          3m
    etcd-1                   1/1       Running   0          3m
    pachd-5655cffbf7-57w4p   1/1       Running   0          3m
```

### Known issues

Problems related to OpenShift deployment are tracked in [issues with the "openshift" label](https://github.com/pachyderm/pachyderm/issues?utf8=%E2%9C%93&q=is%3Aissue+is%3Aopen+label%3Aopenshift).

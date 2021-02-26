package assets

import (
	"fmt"
	"path"
	"strings"

	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/serde"
	apps "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// TODO: Refactor the stateful set setup to better capture the shared functionality between the etcd / postgres setup.
// New / existing features that apply to both should be captured in one place.
// TODO: Move off of kubernetes Deployment object entirely since it is not well suited for stateful applications.
// The primary motivation for this would be to avoid the deadlock that can occur when using a ReadWriteOnce volume mount
// with a kubernetes Deployment.

var (
	// Using our own etcd image for now because there's a fix we need
	// that hasn't been released, and which has been manually applied
	// to the official v3.2.7 release.
	etcdImage = "pachyderm/etcd:v3.3.5"

	etcdHeadlessServiceName = "etcd-headless"
	etcdName                = "etcd"
	etcdVolumeName          = "etcd-volume"
	etcdVolumeClaimName     = "etcd-storage"
	// The storage class name to use when creating a new StorageClass for etcd.
	defaultEtcdStorageClassName = "etcd-storage-class"

	// Cmd used to launch etcd
	etcdCmd = []string{
		"/usr/local/bin/etcd",
		"--listen-client-urls=http://0.0.0.0:2379",
		"--advertise-client-urls=http://0.0.0.0:2379",
		"--data-dir=/var/data/etcd",
		"--auto-compaction-retention=1",
		"--max-txn-ops=10000",
		"--max-request-bytes=52428800",     //50mb
		"--quota-backend-bytes=8589934592", //8gb
	}
)

// EtcdOpts are options that are applicable to etcd.
type EtcdOpts struct {
	Nodes  int
	Volume string

	// CPURequest is the amount of CPU (in cores) we request for each etcd
	// node. If empty, assets.go will choose a default size.
	CPURequest string

	// MemRequest is the amount of memory we request for each etcd node. If
	// empty, assets.go will choose a default size.
	MemRequest string

	// StorageClassName is the name of an existing StorageClass to use when
	// creating a StatefulSet for dynamic etcd storage. If unset, a new
	// StorageClass will be created for the StatefulSet.
	StorageClassName string
}

// WriteEtcdAssets generates all of the etcd-related parts of the kubernetes
// manifest according to the given options and writes it into the given encoder.
func WriteEtcdAssets(encoder serde.Encoder, opts *AssetOpts, objectStoreBackend Backend,
	persistentDiskBackend Backend, volumeSize int,
	hostPath string) error {
	if opts.EtcdOpts.Nodes > 0 && opts.EtcdOpts.Volume != "" {
		return errors.Errorf("only one of --dynamic-etcd-nodes and --static-etcd-volume should be given, but not both")
	}

	// In the dynamic route, we create a storage class which dynamically
	// provisions volumes, and run etcd as a stateful set.
	// In the static route, we create a single volume, a single volume
	// claim, and run etcd as a replication controller with a single node.
	if persistentDiskBackend == LocalBackend {
		if err := encoder.Encode(EtcdDeployment(opts, hostPath)); err != nil {
			return err
		}
	} else if opts.EtcdOpts.Nodes > 0 {
		// Create a StorageClass, if the user didn't provide one.
		if opts.EtcdOpts.StorageClassName == "" {
			sc, err := EtcdStorageClass(opts, persistentDiskBackend)
			if err != nil {
				return err
			}
			if sc != nil {
				if err = encoder.Encode(sc); err != nil {
					return err
				}
			}
		}
		if err := encoder.Encode(EtcdHeadlessService(opts)); err != nil {
			return err
		}
		if err := encoder.Encode(EtcdStatefulSet(opts, persistentDiskBackend, volumeSize)); err != nil {
			return err
		}
	} else if opts.EtcdOpts.Volume != "" {
		volume, err := EtcdVolume(persistentDiskBackend, opts, hostPath, opts.EtcdOpts.Volume, volumeSize)
		if err != nil {
			return err
		}
		if err = encoder.Encode(volume); err != nil {
			return err
		}
		if err = encoder.Encode(EtcdVolumeClaim(volumeSize, opts)); err != nil {
			return err
		}
		if err = encoder.Encode(EtcdDeployment(opts, "")); err != nil {
			return err
		}
	} else {
		return errors.Errorf("unless deploying locally, either --dynamic-etcd-nodes or --static-etcd-volume needs to be provided")
	}
	if err := encoder.Encode(EtcdNodePortService(persistentDiskBackend == LocalBackend, opts)); err != nil {
		return err
	}

	return nil
}

// EtcdDeployment returns an etcd k8s Deployment.
func EtcdDeployment(opts *AssetOpts, hostPath string) *apps.Deployment {
	cpu := resource.MustParse(opts.EtcdOpts.CPURequest)
	mem := resource.MustParse(opts.EtcdOpts.MemRequest)
	var volumes []v1.Volume
	if hostPath == "" {
		volumes = []v1.Volume{
			{
				Name: "etcd-storage",
				VolumeSource: v1.VolumeSource{
					PersistentVolumeClaim: &v1.PersistentVolumeClaimVolumeSource{
						ClaimName: etcdVolumeClaimName,
					},
				},
			},
		}
	} else {
		pathType := v1.HostPathDirectoryOrCreate
		volumes = []v1.Volume{
			{
				Name: "etcd-storage",
				VolumeSource: v1.VolumeSource{
					HostPath: &v1.HostPathVolumeSource{
						Path: path.Join(hostPath, "etcd"),
						Type: &pathType,
					},
				},
			},
		}
	}
	resourceRequirements := v1.ResourceRequirements{
		Requests: v1.ResourceList{
			v1.ResourceCPU:    cpu,
			v1.ResourceMemory: mem,
		},
	}
	if !opts.NoGuaranteed {
		resourceRequirements.Limits = v1.ResourceList{
			v1.ResourceCPU:    cpu,
			v1.ResourceMemory: mem,
		}
	}
	// Don't want to strip the registry out of etcdImage since it's from quay
	// not docker hub.
	image := etcdImage
	if opts.Registry != "" {
		image = AddRegistry(opts.Registry, etcdImage)
	}
	return &apps.Deployment{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Deployment",
			APIVersion: "apps/v1",
		},
		ObjectMeta: objectMeta(etcdName, labels(etcdName), nil, opts.Namespace),
		Spec: apps.DeploymentSpec{
			Replicas: replicas(1),
			Selector: &metav1.LabelSelector{
				MatchLabels: labels(etcdName),
			},
			Template: v1.PodTemplateSpec{
				ObjectMeta: objectMeta(etcdName, labels(etcdName), nil, opts.Namespace),
				Spec: v1.PodSpec{
					Containers: []v1.Container{
						{
							Name:  etcdName,
							Image: image,
							//TODO figure out how to get a cluster of these to talk to each other
							Command: etcdCmd,
							Ports: []v1.ContainerPort{
								{
									ContainerPort: 2379,
									Name:          "client-port",
								},
								{
									ContainerPort: 2380,
									Name:          "peer-port",
								},
							},
							VolumeMounts: []v1.VolumeMount{
								{
									Name:      "etcd-storage",
									MountPath: "/var/data/etcd",
								},
							},
							ImagePullPolicy: "IfNotPresent",
							Resources:       resourceRequirements,
						},
					},
					Volumes:          volumes,
					ImagePullSecrets: imagePullSecrets(opts),
				},
			},
		},
	}
}

// EtcdStorageClass creates a storage class used for dynamic volume
// provisioning.  Currently dynamic volume provisioning only works
// on AWS and GCE.
func EtcdStorageClass(opts *AssetOpts, backend Backend) (interface{}, error) {
	return makeStorageClass(opts, backend, defaultEtcdStorageClassName, labels(etcdName))
}

// EtcdHeadlessService returns a headless etcd service, which is only for DNS
// resolution.
func EtcdHeadlessService(opts *AssetOpts) *v1.Service {
	ports := []v1.ServicePort{
		{
			Name: "peer-port",
			Port: 2380,
		},
	}
	return makeHeadlessService(opts, etcdName, etcdHeadlessServiceName, ports)
}

// EtcdStatefulSet returns a stateful set that manages an etcd cluster
func EtcdStatefulSet(opts *AssetOpts, backend Backend, diskSpace int) interface{} {
	mem := resource.MustParse(opts.EtcdOpts.MemRequest)
	cpu := resource.MustParse(opts.EtcdOpts.CPURequest)
	initialCluster := make([]string, 0, opts.EtcdOpts.Nodes)
	for i := 0; i < opts.EtcdOpts.Nodes; i++ {
		url := fmt.Sprintf("http://etcd-%d.etcd-headless.${NAMESPACE}.svc.cluster.local:2380", i)
		initialCluster = append(initialCluster, fmt.Sprintf("etcd-%d=%s", i, url))
	}
	// Because we need to refer to some environment variables set the by the
	// k8s downward API, we define the command for running etcd here, and then
	// actually run it below via '/bin/sh -c ${CMD}'
	etcdCmd := append(etcdCmd,
		"--listen-peer-urls=http://0.0.0.0:2380",
		"--initial-cluster-token=pach-cluster", // unique ID
		"--initial-advertise-peer-urls=http://${ETCD_NAME}.etcd-headless.${NAMESPACE}.svc.cluster.local:2380",
		"--initial-cluster="+strings.Join(initialCluster, ","),
	)
	for i, str := range etcdCmd {
		etcdCmd[i] = fmt.Sprintf("\"%s\"", str) // quote all arguments, for shell
	}

	var pvcTemplates []interface{}
	switch backend {
	case GoogleBackend, AmazonBackend:
		storageClassName := opts.EtcdOpts.StorageClassName
		if storageClassName == "" {
			storageClassName = defaultEtcdStorageClassName
		}
		pvcTemplates = []interface{}{
			map[string]interface{}{
				"metadata": map[string]interface{}{
					"name":   etcdVolumeClaimName,
					"labels": labels(etcdName),
					"annotations": map[string]string{
						"volume.beta.kubernetes.io/storage-class": storageClassName,
					},
					"namespace": opts.Namespace,
				},
				"spec": map[string]interface{}{
					"resources": map[string]interface{}{
						"requests": map[string]interface{}{
							"storage": resource.MustParse(fmt.Sprintf("%vGi", diskSpace)),
						},
					},
					"accessModes": []string{"ReadWriteOnce"},
				},
			},
		}
	default:
		pvcTemplates = []interface{}{
			map[string]interface{}{
				"metadata": map[string]interface{}{
					"name":      etcdVolumeClaimName,
					"labels":    labels(etcdName),
					"namespace": opts.Namespace,
				},
				"spec": map[string]interface{}{
					"resources": map[string]interface{}{
						"requests": map[string]interface{}{
							"storage": resource.MustParse(fmt.Sprintf("%vGi", diskSpace)),
						},
					},
					"accessModes": []string{"ReadWriteOnce"},
				},
			},
		}
	}
	var imagePullSecrets []map[string]string
	if opts.ImagePullSecret != "" {
		imagePullSecrets = append(imagePullSecrets, map[string]string{"name": opts.ImagePullSecret})
	}
	// As of March 17, 2017, the Kubernetes client does not include structs for
	// Stateful Set, so we generate the kubernetes manifest using raw json.
	// TODO(msteffen): we're now upgrading our kubernetes client, so we should be
	// abe to rewrite this spec using k8s client structs
	image := etcdImage
	if opts.Registry != "" {
		image = AddRegistry(opts.Registry, etcdImage)
	}
	return map[string]interface{}{
		"apiVersion": "apps/v1",
		"kind":       "StatefulSet",
		"metadata": map[string]interface{}{
			"name":      etcdName,
			"labels":    labels(etcdName),
			"namespace": opts.Namespace,
		},
		"spec": map[string]interface{}{
			// Effectively configures a RC
			"serviceName": etcdHeadlessServiceName,
			"replicas":    int(opts.EtcdOpts.Nodes),
			"selector": map[string]interface{}{
				"matchLabels": labels(etcdName),
			},

			// pod template
			"template": map[string]interface{}{
				"metadata": map[string]interface{}{
					"name":      etcdName,
					"labels":    labels(etcdName),
					"namespace": opts.Namespace,
				},
				"spec": map[string]interface{}{
					"imagePullSecrets": imagePullSecrets,
					"containers": []interface{}{
						map[string]interface{}{
							"name":    etcdName,
							"image":   image,
							"command": []string{"/bin/sh", "-c"},
							"args":    []string{strings.Join(etcdCmd, " ")},
							// Use the downward API to pass the pod name to etcd. This sets
							// the etcd-internal name of each node to its pod name.
							"env": []map[string]interface{}{{
								"name": "ETCD_NAME",
								"valueFrom": map[string]interface{}{
									"fieldRef": map[string]interface{}{
										"apiVersion": "v1",
										"fieldPath":  "metadata.name",
									},
								},
							}, {
								"name": "NAMESPACE",
								"valueFrom": map[string]interface{}{
									"fieldRef": map[string]interface{}{
										"apiVersion": "v1",
										"fieldPath":  "metadata.namespace",
									},
								},
							}},
							"ports": []interface{}{
								map[string]interface{}{
									"containerPort": 2379,
									"name":          "client-port",
								},
								map[string]interface{}{
									"containerPort": 2380,
									"name":          "peer-port",
								},
							},
							"volumeMounts": []interface{}{
								map[string]interface{}{
									"name":      etcdVolumeClaimName,
									"mountPath": "/var/data/etcd",
								},
							},
							"imagePullPolicy": "IfNotPresent",
							"resources": map[string]interface{}{
								"requests": map[string]interface{}{
									string(v1.ResourceCPU):    cpu.String(),
									string(v1.ResourceMemory): mem.String(),
								},
							},
						},
					},
				},
			},
			"volumeClaimTemplates": pvcTemplates,
		},
	}
}

// EtcdVolume creates a persistent volume backed by a volume with name "name"
func EtcdVolume(persistentDiskBackend Backend, opts *AssetOpts,
	hostPath string, name string, size int) (*v1.PersistentVolume, error) {
	return makePersistentVolume(opts, persistentDiskBackend, hostPath, name, size, etcdVolumeName, labels(etcdName))
}

// EtcdVolumeClaim creates a persistent volume claim of 'size' GB.
//
// Note that if you're controlling Etcd with a Stateful Set, this is
// unnecessary (the stateful set controller will create PVCs automatically).
func EtcdVolumeClaim(size int, opts *AssetOpts) *v1.PersistentVolumeClaim {
	return makeVolumeClaim(opts, size, etcdVolumeName, etcdVolumeClaimName, labels(etcdName))
}

// EtcdNodePortService returns a NodePort etcd service. This will let non-etcd
// pods talk to etcd
func EtcdNodePortService(local bool, opts *AssetOpts) *v1.Service {
	var clientNodePort int32
	if local {
		clientNodePort = 32379
	}
	return &v1.Service{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Service",
			APIVersion: "v1",
		},
		ObjectMeta: objectMeta(etcdName, labels(etcdName), nil, opts.Namespace),
		Spec: v1.ServiceSpec{
			Type: v1.ServiceTypeNodePort,
			Selector: map[string]string{
				"app": etcdName,
			},
			Ports: []v1.ServicePort{
				{
					Port:     2379,
					Name:     "client-port",
					NodePort: clientNodePort,
				},
			},
		},
	}
}

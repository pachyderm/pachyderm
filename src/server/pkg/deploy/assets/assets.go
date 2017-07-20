package assets

import (
	"fmt"
	"io"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/pachyderm/pachyderm/src/server/pfs/server"
	"github.com/pachyderm/pachyderm/src/server/pkg/deploy"
	"github.com/ugorji/go/codec"
	"k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/api/resource"
	"k8s.io/kubernetes/pkg/api/unversioned"
	// k8s.io/kubernetes/pkg/api/v1 is very similar to
	// "k8s.io/kubernetes/pkg/api" above, we import both because services need
	// to use v1 otherwise they get empty fields that make them invalid to
	// kubectl, we need the api version to interact correctly with deployments
	"k8s.io/kubernetes/pkg/api/v1"
	"k8s.io/kubernetes/pkg/apis/extensions"
)

var (
	suite                   = "pachyderm"
	pachdImage              = "pachyderm/pachd"
	etcdImage               = "quay.io/coreos/etcd:v3.1.4"
	serviceAccountName      = "pachyderm"
	etcdHeadlessServiceName = "etcd-headless"
	etcdName                = "etcd"
	etcdVolumeName          = "etcd-volume"
	etcdVolumeClaimName     = "etcd-storage"
	etcdStorageClassName    = "etcd-storage-class"
	dashName                = "dash"
	dashImage               = "pachyderm/dash"
	grpcProxyName           = "grpc-proxy"
	grpcProxyImage          = "pachyderm/grpc-proxy:0.3.0"
	pachdName               = "pachd"
	minioSecretName         = "minio-secret"
	amazonSecretName        = "amazon-secret"
	googleSecretName        = "google-secret"
	microsoftSecretName     = "microsoft-secret"
	trueVal                 = true
	jsonEncoderHandle       = &codec.JsonHandle{
		BasicHandle: codec.BasicHandle{
			EncodeOptions: codec.EncodeOptions{Canonical: true},
		},
		Indent: 2,
	}
)

type backend int

const (
	localBackend backend = iota
	amazonBackend
	googleBackend
	microsoftBackend
	minioBackend
	s3CustomArgs = 6
)

// AssetOpts are options that are applicable to all the asset types.
type AssetOpts struct {
	PachdShards uint64
	Version     string
	LogLevel    string
	Metrics     bool
	Dynamic     bool
	EtcdNodes   int
	EtcdVolume  string
	EnableDash  bool
	DashOnly    bool
	DashImage   string

	// BlockCacheSize is the amount of memory each PachD node allocates towards
	// its cache of PFS blocks. If empty, assets.go will choose a default size.
	BlockCacheSize string

	// PachdCPURequest is the amount of CPU we request for each pachd node. If
	// empty, assets.go will choose a default size.
	PachdCPURequest string

	// PachdNonCacheMemRequest is the amount of memory we request for each
	// pachd node in addition to BlockCacheSize. If empty, assets.go will choose
	// a default size.
	PachdNonCacheMemRequest string

	// EtcdCPURequest is the amount of CPU (in cores) we request for each etcd
	// node. If empty, assets.go will choose a default size.
	EtcdCPURequest string

	// EtcdMemRequest is the amount of memory we request for each etcd node. If
	// empty, assets.go will choose a default size.
	EtcdMemRequest string
}

// fillDefaultResourceRequests sets any of:
//   opts.BlockCacheSize
//   opts.PachdNonCacheMemRequest
//   opts.PachdCPURequest
//   opts.EtcdCPURequest
//   opts.EtcdMemRequest
// that are unset in 'opts' to the appropriate default ('persistentDiskBackend'
// just used to determine if this is a local deployment, and if so, make the
// resource requests smaller)
func fillDefaultResourceRequests(opts *AssetOpts, persistentDiskBackend backend) {
	if persistentDiskBackend == localBackend {
		// For local deployments, we set the resource requirements and cache sizes
		// low so that pachyderm clusters will fit inside e.g. minikube or travis
		if opts.BlockCacheSize == "" {
			opts.BlockCacheSize = "256M"
		}
		if opts.PachdNonCacheMemRequest == "" {
			opts.PachdNonCacheMemRequest = "256M"
		}
		if opts.PachdCPURequest == "" {
			opts.PachdCPURequest = "0.25"
		}

		if opts.EtcdMemRequest == "" {
			opts.EtcdMemRequest = "256M"
		}
		if opts.EtcdCPURequest == "" {
			opts.EtcdCPURequest = "0.25"
		}
	} else {
		// For non-local deployments, we set the resource requirements and cache
		// sizes higher, so that the cluster is stable and performant
		if opts.BlockCacheSize == "" {
			opts.BlockCacheSize = "1G"
		}
		if opts.PachdNonCacheMemRequest == "" {
			opts.PachdNonCacheMemRequest = "2G"
		}
		if opts.PachdCPURequest == "" {
			opts.PachdCPURequest = "1"
		}

		if opts.EtcdMemRequest == "" {
			opts.EtcdMemRequest = "2G"
		}
		if opts.EtcdCPURequest == "" {
			opts.EtcdCPURequest = "1"
		}
	}
}

// ServiceAccount returns a kubernetes service account for use with Pachyderm.
func ServiceAccount() *api.ServiceAccount {
	return &api.ServiceAccount{
		TypeMeta: unversioned.TypeMeta{
			Kind:       "ServiceAccount",
			APIVersion: "v1",
		},
		ObjectMeta: api.ObjectMeta{
			Name:   serviceAccountName,
			Labels: labels(""),
		},
	}
}

// GetSecretVolumeAndMount returns a properly configured Volume and
// VolumeMount object given a backend.  The backend needs to be one of the
// constants defined in pfs/server.
func GetSecretVolumeAndMount(backend string) (api.Volume, api.VolumeMount, error) {
	switch backend {
	case server.MinioBackendEnvVar:
		return api.Volume{
				Name: minioSecretName,
				VolumeSource: api.VolumeSource{
					Secret: &api.SecretVolumeSource{
						SecretName: minioSecretName,
					},
				},
			}, api.VolumeMount{
				Name:      minioSecretName,
				MountPath: "/" + minioSecretName,
			}, nil
	case server.AmazonBackendEnvVar:
		return api.Volume{
				Name: amazonSecretName,
				VolumeSource: api.VolumeSource{
					Secret: &api.SecretVolumeSource{
						SecretName: amazonSecretName,
					},
				},
			}, api.VolumeMount{
				Name:      amazonSecretName,
				MountPath: "/" + amazonSecretName,
			}, nil
	case server.GoogleBackendEnvVar:
		return api.Volume{
				Name: googleSecretName,
				VolumeSource: api.VolumeSource{
					Secret: &api.SecretVolumeSource{
						SecretName: googleSecretName,
					},
				},
			}, api.VolumeMount{
				Name:      googleSecretName,
				MountPath: "/" + googleSecretName,
			}, nil
	case server.MicrosoftBackendEnvVar:
		return api.Volume{
				Name: microsoftSecretName,
				VolumeSource: api.VolumeSource{
					Secret: &api.SecretVolumeSource{
						SecretName: microsoftSecretName,
					},
				},
			}, api.VolumeMount{
				Name:      microsoftSecretName,
				MountPath: "/" + microsoftSecretName,
			}, nil
	}
	return api.Volume{}, api.VolumeMount{}, fmt.Errorf("not found")
}

// PachdDeployment returns a pachd k8s Deployment.
func PachdDeployment(opts *AssetOpts, objectStoreBackend backend, hostPath string) *extensions.Deployment {
	mem := resource.MustParse(opts.BlockCacheSize)
	mem.Add(resource.MustParse(opts.PachdNonCacheMemRequest))
	cpu := resource.MustParse(opts.PachdCPURequest)
	image := pachdImage
	if opts.Version != "" {
		image += ":" + opts.Version
	}
	// we turn metrics off if we dont have a static version
	// this prevents dev clusters from reporting metrics
	if opts.Version == deploy.DevVersionTag {
		opts.Metrics = false
	}
	volumes := []api.Volume{
		{
			Name: "pach-disk",
		},
	}
	volumeMounts := []api.VolumeMount{
		{
			Name:      "pach-disk",
			MountPath: "/pach",
		},
	}
	var backendEnvVar string
	var storageHostPath string
	switch objectStoreBackend {
	case localBackend:
		storageHostPath = filepath.Join(hostPath, "pachd")
		volumes[0].HostPath = &api.HostPathVolumeSource{
			Path: storageHostPath,
		}
	case minioBackend:
		backendEnvVar = server.MinioBackendEnvVar
	case amazonBackend:
		backendEnvVar = server.AmazonBackendEnvVar
	case googleBackend:
		backendEnvVar = server.GoogleBackendEnvVar
	case microsoftBackend:
		backendEnvVar = server.MicrosoftBackendEnvVar
	}
	volume, mount, err := GetSecretVolumeAndMount(backendEnvVar)
	if err == nil {
		volumes = append(volumes, volume)
		volumeMounts = append(volumeMounts, mount)
	}
	return &extensions.Deployment{
		TypeMeta: unversioned.TypeMeta{
			Kind:       "Deployment",
			APIVersion: "extensions/v1beta1",
		},
		ObjectMeta: api.ObjectMeta{
			Name:   pachdName,
			Labels: labels(pachdName),
		},
		Spec: extensions.DeploymentSpec{
			Replicas: 1,
			Selector: &unversioned.LabelSelector{
				MatchLabels: labels(pachdName),
			},
			Template: api.PodTemplateSpec{
				ObjectMeta: api.ObjectMeta{
					Name:   pachdName,
					Labels: labels(pachdName),
				},
				Spec: api.PodSpec{
					Containers: []api.Container{
						{
							Name:  pachdName,
							Image: image,
							Env: []api.EnvVar{
								{
									Name:  "PACH_ROOT",
									Value: "/pach",
								},
								{
									Name:  "NUM_SHARDS",
									Value: fmt.Sprintf("%d", opts.PachdShards),
								},
								{
									Name:  "STORAGE_BACKEND",
									Value: backendEnvVar,
								},
								{
									Name:  "STORAGE_HOST_PATH",
									Value: storageHostPath,
								},
								{
									Name: "PACHD_POD_NAMESPACE",
									ValueFrom: &api.EnvVarSource{
										FieldRef: &api.ObjectFieldSelector{
											APIVersion: "v1",
											FieldPath:  "metadata.namespace",
										},
									},
								},
								{
									Name:  "WORKER_IMAGE",
									Value: fmt.Sprintf("pachyderm/worker:%s", opts.Version),
								},
								{
									Name:  "WORKER_SIDECAR_IMAGE",
									Value: fmt.Sprintf("pachyderm/pachd:%s", opts.Version),
								},
								{
									Name:  "WORKER_IMAGE_PULL_POLICY",
									Value: "IfNotPresent",
								},
								{
									Name:  "PACHD_VERSION",
									Value: opts.Version,
								},
								{
									Name:  "METRICS",
									Value: strconv.FormatBool(opts.Metrics),
								},
								{
									Name:  "LOG_LEVEL",
									Value: opts.LogLevel,
								},
								{
									Name:  "BLOCK_CACHE_BYTES",
									Value: opts.BlockCacheSize,
								},
							},
							Ports: []api.ContainerPort{
								{
									ContainerPort: 650,
									Protocol:      "TCP",
									Name:          "api-grpc-port",
								},
								{
									ContainerPort: 651,
									Name:          "trace-port",
								},
								{
									ContainerPort: server.HTTPPort,
									Protocol:      "TCP",
									Name:          "api-http-port",
								},
							},
							VolumeMounts: volumeMounts,
							SecurityContext: &api.SecurityContext{
								Privileged: &trueVal, // god is this dumb
							},
							ImagePullPolicy: "IfNotPresent",
							Resources: api.ResourceRequirements{
								Requests: api.ResourceList{
									api.ResourceCPU:    cpu,
									api.ResourceMemory: mem,
								},
							},
						},
					},
					ServiceAccountName: serviceAccountName,
					Volumes:            volumes,
				},
			},
		},
	}
}

// PachdService returns a pachd service.
func PachdService() *v1.Service {
	return &v1.Service{
		TypeMeta: unversioned.TypeMeta{
			Kind:       "Service",
			APIVersion: "v1",
		},
		ObjectMeta: v1.ObjectMeta{
			Name:   pachdName,
			Labels: labels(pachdName),
		},
		Spec: v1.ServiceSpec{
			Type: v1.ServiceTypeNodePort,
			Selector: map[string]string{
				"app": pachdName,
			},
			Ports: []v1.ServicePort{
				{
					Port:     650,
					Name:     "api-grpc-port",
					NodePort: 30650,
				},
				{
					Port:     651,
					Name:     "trace-port",
					NodePort: 30651,
				},
				{
					Port:     server.HTTPPort,
					Name:     "api-http-port",
					NodePort: 30000 + server.HTTPPort,
				},
			},
		},
	}
}

// EtcdDeployment returns an etcd k8s Deployment.
func EtcdDeployment(opts *AssetOpts, hostPath string) *extensions.Deployment {
	cpu := resource.MustParse(opts.EtcdCPURequest)
	mem := resource.MustParse(opts.EtcdMemRequest)
	var volumes []api.Volume
	if hostPath == "" {
		volumes = []api.Volume{
			{
				Name: "etcd-storage",
				VolumeSource: api.VolumeSource{
					PersistentVolumeClaim: &api.PersistentVolumeClaimVolumeSource{
						ClaimName: etcdVolumeClaimName,
					},
				},
			},
		}
	} else {
		volumes = []api.Volume{
			{
				Name: "etcd-storage",
				VolumeSource: api.VolumeSource{
					HostPath: &api.HostPathVolumeSource{
						Path: filepath.Join(hostPath, "etcd"),
					},
				},
			},
		}
	}
	return &extensions.Deployment{
		TypeMeta: unversioned.TypeMeta{
			Kind:       "Deployment",
			APIVersion: "extensions/v1beta1",
		},
		ObjectMeta: api.ObjectMeta{
			Name:   etcdName,
			Labels: labels(etcdName),
		},
		Spec: extensions.DeploymentSpec{
			Replicas: 1,
			Selector: &unversioned.LabelSelector{
				MatchLabels: labels(etcdName),
			},
			Template: api.PodTemplateSpec{
				ObjectMeta: api.ObjectMeta{
					Name:   etcdName,
					Labels: labels(etcdName),
				},
				Spec: api.PodSpec{
					Containers: []api.Container{
						{
							Name:  etcdName,
							Image: etcdImage,
							//TODO figure out how to get a cluster of these to talk to each other
							Command: []string{
								"/usr/local/bin/etcd",
								"--listen-client-urls=http://0.0.0.0:2379",
								"--advertise-client-urls=http://0.0.0.0:2379",
								"--data-dir=/var/data/etcd",
								"--auto-compaction-retention=1",
							},
							Ports: []api.ContainerPort{
								{
									ContainerPort: 2379,
									Name:          "client-port",
								},
								{
									ContainerPort: 2380,
									Name:          "peer-port",
								},
							},
							VolumeMounts: []api.VolumeMount{
								{
									Name:      "etcd-storage",
									MountPath: "/var/data/etcd",
								},
							},
							ImagePullPolicy: "IfNotPresent",
							Resources: api.ResourceRequirements{
								Requests: api.ResourceList{
									api.ResourceCPU:    cpu,
									api.ResourceMemory: mem,
								},
							},
						},
					},
					Volumes: volumes,
				},
			},
		},
	}
}

// EtcdStorageClass creates a storage class used for dynamic volume
// provisioning.  Currently dynamic volume provisioning only works
// on AWS and GCE.
func EtcdStorageClass(backend backend) (interface{}, error) {
	sc := map[string]interface{}{
		"apiVersion": "storage.k8s.io/v1beta1",
		"kind":       "StorageClass",
		"metadata": map[string]interface{}{
			"name":   etcdStorageClassName,
			"labels": labels(etcdName),
		},
	}
	switch backend {
	case googleBackend:
		sc["provisioner"] = "kubernetes.io/gce-pd"
		sc["parameters"] = map[string]string{
			"type": "pd-ssd",
		}
	case amazonBackend:
		sc["provisioner"] = "kubernetes.io/aws-ebs"
		sc["parameters"] = map[string]string{
			"type": "gp2",
		}
	default:
		return nil, nil
	}
	return sc, nil
}

// EtcdVolume creates a persistent volume backed by a volume with name "name"
func EtcdVolume(persistentDiskBackend backend, opts *AssetOpts,
	hostPath string, name string, size int) (*api.PersistentVolume, error) {
	spec := &api.PersistentVolume{
		TypeMeta: unversioned.TypeMeta{
			Kind:       "PersistentVolume",
			APIVersion: "v1",
		},
		ObjectMeta: api.ObjectMeta{
			Name:   etcdVolumeName,
			Labels: labels(etcdName),
		},
		Spec: api.PersistentVolumeSpec{
			Capacity: map[api.ResourceName]resource.Quantity{
				"storage": resource.MustParse(fmt.Sprintf("%vGi", size)),
			},
			AccessModes:                   []api.PersistentVolumeAccessMode{api.ReadWriteOnce},
			PersistentVolumeReclaimPolicy: api.PersistentVolumeReclaimRetain,
		},
	}

	switch persistentDiskBackend {
	case amazonBackend:
		spec.Spec.PersistentVolumeSource = api.PersistentVolumeSource{
			AWSElasticBlockStore: &api.AWSElasticBlockStoreVolumeSource{
				FSType:   "ext4",
				VolumeID: name,
			},
		}
	case googleBackend:
		spec.Spec.PersistentVolumeSource = api.PersistentVolumeSource{
			GCEPersistentDisk: &api.GCEPersistentDiskVolumeSource{
				FSType: "ext4",
				PDName: name,
			},
		}
	case microsoftBackend:
		dataDiskURI := name
		split := strings.Split(name, "/")
		diskName := split[len(split)-1]

		spec.Spec.PersistentVolumeSource = api.PersistentVolumeSource{
			AzureDisk: &api.AzureDiskVolumeSource{
				DiskName:    diskName,
				DataDiskURI: dataDiskURI,
			},
		}
	case minioBackend:
		fallthrough
	case localBackend:
		spec.Spec.PersistentVolumeSource = api.PersistentVolumeSource{
			HostPath: &api.HostPathVolumeSource{
				Path: filepath.Join(hostPath, "etcd"),
			},
		}
	default:
		return nil, fmt.Errorf("cannot generate volume spec for unknown backend \"%v\"", persistentDiskBackend)
	}
	return spec, nil
}

// EtcdVolumeClaim creates a persistent volume claim of 'size' GB.
//
// Note that if you're controlling Etcd with a Stateful Set, this is
// unneccessary (the stateful set controller will create PVCs automatically).
func EtcdVolumeClaim(size int) *api.PersistentVolumeClaim {
	return &api.PersistentVolumeClaim{
		TypeMeta: unversioned.TypeMeta{
			Kind:       "PersistentVolumeClaim",
			APIVersion: "v1",
		},
		ObjectMeta: api.ObjectMeta{
			Name:   etcdVolumeClaimName,
			Labels: labels(etcdName),
		},
		Spec: api.PersistentVolumeClaimSpec{
			Resources: api.ResourceRequirements{
				Requests: map[api.ResourceName]resource.Quantity{
					"storage": resource.MustParse(fmt.Sprintf("%vGi", size)),
				},
			},
			AccessModes: []api.PersistentVolumeAccessMode{api.ReadWriteOnce},
			VolumeName:  etcdVolumeName,
		},
	}
}

// EtcdNodePortService returns a NodePort etcd service. This will let non-etcd
// pods talk to etcd
func EtcdNodePortService(local bool) *v1.Service {
	var clientNodePort int32
	if local {
		clientNodePort = 32379
	}
	return &v1.Service{
		TypeMeta: unversioned.TypeMeta{
			Kind:       "Service",
			APIVersion: "v1",
		},
		ObjectMeta: v1.ObjectMeta{
			Name:   etcdName,
			Labels: labels(etcdName),
		},
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

// EtcdHeadlessService returns a headless etcd service, which is only for DNS
// resolution.
func EtcdHeadlessService() *v1.Service {
	return &v1.Service{
		TypeMeta: unversioned.TypeMeta{
			Kind:       "Service",
			APIVersion: "v1",
		},
		ObjectMeta: v1.ObjectMeta{
			Name:   etcdHeadlessServiceName,
			Labels: labels(etcdName),
		},
		Spec: v1.ServiceSpec{
			Selector: map[string]string{
				"app": etcdName,
			},
			ClusterIP: "None",
			Ports: []v1.ServicePort{
				{
					Name: "peer-port",
					Port: 2380,
				},
			},
		},
	}
}

// EtcdStatefulSet returns a stateful set that manages an etcd cluster
func EtcdStatefulSet(opts *AssetOpts, backend backend, diskSpace int) interface{} {
	mem := resource.MustParse(opts.EtcdMemRequest)
	cpu := resource.MustParse(opts.EtcdCPURequest)
	initialCluster := make([]string, 0, opts.EtcdNodes)
	for i := 0; i < opts.EtcdNodes; i++ {
		url := fmt.Sprintf("http://etcd-%d.etcd-headless.${NAMESPACE}.svc.cluster.local:2380", i)
		initialCluster = append(initialCluster, fmt.Sprintf("etcd-%d=%s", i, url))
	}
	// Because we need to refer to some environment variables set the by the
	// k8s downward API, we define the command for running etcd here, and then
	// actually run it below via '/bin/sh -c ${CMD}'
	etcdCmd := []string{
		"/usr/local/bin/etcd",
		"--listen-client-urls=http://0.0.0.0:2379",
		"--advertise-client-urls=http://0.0.0.0:2379",
		"--listen-peer-urls=http://0.0.0.0:2380",
		"--data-dir=/var/data/etcd",
		"--initial-cluster-token=pach-cluster", // unique ID
		"--initial-advertise-peer-urls=http://${ETCD_NAME}.etcd-headless.${NAMESPACE}.svc.cluster.local:2380",
		"--initial-cluster=" + strings.Join(initialCluster, ","),
		"--auto-compaction-retention=1",
	}
	for i, str := range etcdCmd {
		etcdCmd[i] = fmt.Sprintf("\"%s\"", str) // quote all arguments, for shell
	}

	var pvcTemplates []interface{}
	switch backend {
	case googleBackend, amazonBackend:
		pvcTemplates = []interface{}{
			map[string]interface{}{
				"metadata": map[string]interface{}{
					"name":   etcdVolumeClaimName,
					"labels": labels(etcdName),
					"annotations": map[string]string{
						"volume.beta.kubernetes.io/storage-class": etcdStorageClassName,
					},
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
					"name":   etcdVolumeClaimName,
					"labels": labels(etcdName),
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

	// As of March 17, 2017, the Kubernetes client does not include structs for
	// Stateful Set, so we generate the kubernetes manifest using raw json.
	return map[string]interface{}{
		"apiVersion": "apps/v1beta1",
		"kind":       "StatefulSet",
		"metadata": map[string]interface{}{
			"name":   etcdName,
			"labels": labels(etcdName),
		},
		"spec": map[string]interface{}{
			// Effectively configures a RC
			"serviceName": etcdHeadlessServiceName,
			"replicas":    int(opts.EtcdNodes),
			"selector": map[string]interface{}{
				"matchLabels": labels(etcdName),
			},

			// pod template
			"template": map[string]interface{}{
				"metadata": map[string]interface{}{
					"name":   etcdName,
					"labels": labels(etcdName),
				},
				"spec": map[string]interface{}{
					"containers": []interface{}{
						map[string]interface{}{
							"name":    etcdName,
							"image":   etcdImage,
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
									string(api.ResourceCPU):    cpu.String(),
									string(api.ResourceMemory): mem.String(),
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

// DashDeployment creates a Deployment for the pachyderm dashboard.
func DashDeployment(dashImage string) *extensions.Deployment {
	return &extensions.Deployment{
		TypeMeta: unversioned.TypeMeta{
			Kind:       "Deployment",
			APIVersion: "extensions/v1beta1",
		},
		ObjectMeta: api.ObjectMeta{
			Name:   dashName,
			Labels: labels(dashName),
		},
		Spec: extensions.DeploymentSpec{
			Selector: &unversioned.LabelSelector{
				MatchLabels: labels(dashName),
			},
			Template: api.PodTemplateSpec{
				ObjectMeta: api.ObjectMeta{
					Name:   dashName,
					Labels: labels(dashName),
				},
				Spec: api.PodSpec{
					Containers: []api.Container{
						{
							Name:  dashName,
							Image: dashImage,
							Ports: []api.ContainerPort{
								{
									ContainerPort: 8080,
									Name:          "dash-http",
								},
							},
							ImagePullPolicy: "IfNotPresent",
						},
						{
							Name:  grpcProxyName,
							Image: grpcProxyImage,
							Ports: []api.ContainerPort{
								{
									ContainerPort: 8081,
									Name:          "grpc-proxy-http",
								},
							},
							ImagePullPolicy: "IfNotPresent",
						},
					},
				},
			},
		},
	}
}

// DashService creates a Service for the pachyderm dashboard.
func DashService() *v1.Service {
	return &v1.Service{
		TypeMeta: unversioned.TypeMeta{
			Kind:       "Service",
			APIVersion: "v1",
		},
		ObjectMeta: v1.ObjectMeta{
			Name:   dashName,
			Labels: labels(dashName),
		},
		Spec: v1.ServiceSpec{
			Type:     v1.ServiceTypeNodePort,
			Selector: labels(dashName),
			Ports: []v1.ServicePort{
				{
					Port:     8080,
					Name:     "dash-http",
					NodePort: 30080,
				},
				{
					Port:     8081,
					Name:     "grpc-proxy-http",
					NodePort: 30081,
				},
			},
		},
	}
}

// MinioSecret creates an amazon secret with the following parameters:
//   bucket - S3 bucket name
//   id     - S3 access key id
//   secret - S3 secret access key
//   endpoint  - S3 compatible endpoint
//   secure - set to true for a secure connection.
func MinioSecret(bucket string, id string, secret string, endpoint string, secure bool) *api.Secret {
	secureV := "0"
	if secure {
		secureV = "1"
	}
	return &api.Secret{
		TypeMeta: unversioned.TypeMeta{
			Kind:       "Secret",
			APIVersion: "v1",
		},
		ObjectMeta: api.ObjectMeta{
			Name:   minioSecretName,
			Labels: labels(minioSecretName),
		},
		Data: map[string][]byte{
			"bucket":   []byte(bucket),
			"id":       []byte(id),
			"secret":   []byte(secret),
			"endpoint": []byte(endpoint),
			"secure":   []byte(secureV),
		},
	}
}

// AmazonSecret creates an amazon secret with the following parameters:
//   bucket - S3 bucket name
//   id     - AWS access key id
//   secret - AWS secret access key
//   token  - AWS access token
//   region - AWS region
func AmazonSecret(bucket string, distribution string, id string, secret string, token string, region string) *api.Secret {
	return &api.Secret{
		TypeMeta: unversioned.TypeMeta{
			Kind:       "Secret",
			APIVersion: "v1",
		},
		ObjectMeta: api.ObjectMeta{
			Name:   amazonSecretName,
			Labels: labels(amazonSecretName),
		},
		Data: map[string][]byte{
			"bucket":       []byte(bucket),
			"distribution": []byte(distribution),
			"id":           []byte(id),
			"secret":       []byte(secret),
			"token":        []byte(token),
			"region":       []byte(region),
		},
	}
}

// GoogleSecret creates a google secret with a bucket name.
func GoogleSecret(bucket string) *api.Secret {
	return &api.Secret{
		TypeMeta: unversioned.TypeMeta{
			Kind:       "Secret",
			APIVersion: "v1",
		},
		ObjectMeta: api.ObjectMeta{
			Name:   googleSecretName,
			Labels: labels(googleSecretName),
		},
		Data: map[string][]byte{
			"bucket": []byte(bucket),
		},
	}
}

// MicrosoftSecret creates a microsoft secret with following parameters:
//   container - Azure blob container
//   id    	   - Azure storage account name
//   secret    - Azure storage account key
func MicrosoftSecret(container string, id string, secret string) *api.Secret {
	return &api.Secret{
		TypeMeta: unversioned.TypeMeta{
			Kind:       "Secret",
			APIVersion: "v1",
		},
		ObjectMeta: api.ObjectMeta{
			Name:   microsoftSecretName,
			Labels: labels(microsoftSecretName),
		},
		Data: map[string][]byte{
			"container": []byte(container),
			"id":        []byte(id),
			"secret":    []byte(secret),
		},
	}
}

// WriteDashboardAssets writes the k8s config for deploying the Pachyderm
// dashboard to 'w'
func WriteDashboardAssets(w io.Writer, opts *AssetOpts) {
	encoder := codec.NewEncoder(w, jsonEncoderHandle)
	DashService().CodecEncodeSelf(encoder)
	fmt.Fprintf(w, "\n")
	DashDeployment(opts.DashImage).CodecEncodeSelf(encoder)
	fmt.Fprintf(w, "\n")
}

// WriteAssets writes the assets to w.
func WriteAssets(w io.Writer, opts *AssetOpts, objectStoreBackend backend,
	persistentDiskBackend backend, volumeSize int,
	hostPath string) error {
	// If either backend is "local", both must be "local"
	if (persistentDiskBackend == localBackend || objectStoreBackend == localBackend) &&
		persistentDiskBackend != objectStoreBackend {
		return fmt.Errorf("if either persistentDiskBackend or objectStoreBackend "+
			"is \"local\", both must be \"local\", but persistentDiskBackend==%d, \n"+
			"and objectStoreBackend==%d", persistentDiskBackend, objectStoreBackend)
	}
	fillDefaultResourceRequests(opts, persistentDiskBackend)
	if opts.DashOnly {
		WriteDashboardAssets(w, opts)
		return nil
	}
	encoder := codec.NewEncoder(w, jsonEncoderHandle)

	ServiceAccount().CodecEncodeSelf(encoder)
	fmt.Fprintf(w, "\n")

	if opts.EtcdNodes > 0 && opts.EtcdVolume != "" {
		return fmt.Errorf("only one of --dynamic-etcd-nodes and --static-etcd-volume should be given, but not both")
	}

	// In the dynamic route, we create a storage class which dynamically
	// provisions volumes, and run etcd as a statful set.
	// In the static route, we create a single volume, a single volume
	// claim, and run etcd as a replication controller with a single node.
	if objectStoreBackend == localBackend {
		EtcdDeployment(opts, hostPath).CodecEncodeSelf(encoder)
		fmt.Fprintf(w, "\n")
	} else if opts.EtcdNodes > 0 {
		sc, err := EtcdStorageClass(persistentDiskBackend)
		if err != nil {
			return err
		}
		if sc != nil {
			encoder.Encode(sc)
			fmt.Fprintf(w, "\n")
		}
		EtcdHeadlessService().CodecEncodeSelf(encoder)
		fmt.Fprintf(w, "\n")
		encoder.Encode(EtcdStatefulSet(opts, persistentDiskBackend, volumeSize))
		fmt.Fprintf(w, "\n")
	} else if opts.EtcdVolume != "" || persistentDiskBackend == localBackend {
		volume, err := EtcdVolume(persistentDiskBackend, opts, hostPath, opts.EtcdVolume, volumeSize)
		if err != nil {
			return err
		}
		volume.CodecEncodeSelf(encoder)
		fmt.Fprintf(w, "\n")
		EtcdVolumeClaim(volumeSize).CodecEncodeSelf(encoder)
		fmt.Fprintf(w, "\n")
		EtcdDeployment(opts, "").CodecEncodeSelf(encoder)
		fmt.Fprintf(w, "\n")
	} else {
		return fmt.Errorf("unless deploying locally, either --dynamic-etcd-nodes or --static-etcd-volume needs to be provided")
	}
	EtcdNodePortService(objectStoreBackend == localBackend).CodecEncodeSelf(encoder)
	fmt.Fprintf(w, "\n")

	PachdService().CodecEncodeSelf(encoder)
	fmt.Fprintf(w, "\n")
	PachdDeployment(opts, objectStoreBackend, hostPath).CodecEncodeSelf(encoder)
	fmt.Fprintf(w, "\n")
	if opts.EnableDash {
		WriteDashboardAssets(w, opts)
	}
	return nil
}

// WriteLocalAssets writes assets to a local backend.
func WriteLocalAssets(w io.Writer, opts *AssetOpts, hostPath string) error {
	return WriteAssets(w, opts, localBackend, localBackend, 1 /* = volume size (gb) */, hostPath)
}

// WriteCustomAssets writes assets to a custom combination of object-store and persistent disk.
func WriteCustomAssets(w io.Writer, opts *AssetOpts, args []string, objectStoreBackend string,
	persistentDiskBackend string, secure bool) error {
	switch objectStoreBackend {
	case "s3":
		if len(args) != s3CustomArgs {
			return fmt.Errorf("Expected %d arguments for disk+s3 backend", s3CustomArgs)
		}
		volumeSize, err := strconv.Atoi(args[1])
		if err != nil {
			return fmt.Errorf("volume size needs to be an integer; instead got %v", args[1])
		}
		switch persistentDiskBackend {
		case "aws":
			if err := WriteAssets(w, opts, minioBackend, amazonBackend, volumeSize, ""); err != nil {
				return err
			}
		case "google":
			if err := WriteAssets(w, opts, minioBackend, googleBackend, volumeSize, ""); err != nil {
				return err
			}
		case "azure":
			if err := WriteAssets(w, opts, minioBackend, microsoftBackend, volumeSize, ""); err != nil {
				return err
			}
		default:
			return fmt.Errorf("Did not recognize the choice of persistent-disk")
		}
		encoder := codec.NewEncoder(w, jsonEncoderHandle)
		MinioSecret(args[2], args[3], args[4], args[5], secure).CodecEncodeSelf(encoder)
		fmt.Fprintf(w, "\n")
		return nil
	default:
		return fmt.Errorf("Did not recognize the choice of object-store")
	}
}

// WriteAmazonAssets writes assets to an amazon backend.
func WriteAmazonAssets(w io.Writer, opts *AssetOpts, bucket string, id string, secret string,
	token string, region string, volumeSize int, distribution string) error {
	if err := WriteAssets(w, opts, amazonBackend, amazonBackend, volumeSize, ""); err != nil {
		return err
	}
	encoder := codec.NewEncoder(w, jsonEncoderHandle)
	AmazonSecret(bucket, distribution, id, secret, token, region).CodecEncodeSelf(encoder)
	fmt.Fprintf(w, "\n")
	return nil
}

// WriteGoogleAssets writes assets to a google backend.
func WriteGoogleAssets(w io.Writer, opts *AssetOpts, bucket string, volumeSize int) error {
	if err := WriteAssets(w, opts, googleBackend, googleBackend, volumeSize, ""); err != nil {
		return err
	}
	encoder := codec.NewEncoder(w, jsonEncoderHandle)
	GoogleSecret(bucket).CodecEncodeSelf(encoder)
	fmt.Fprintf(w, "\n")
	return nil
}

// WriteMicrosoftAssets writes assets to a microsoft backend
func WriteMicrosoftAssets(w io.Writer, opts *AssetOpts, container string, id string, secret string, volumeSize int) error {
	if err := WriteAssets(w, opts, microsoftBackend, microsoftBackend, volumeSize, ""); err != nil {
		return err
	}
	encoder := codec.NewEncoder(w, jsonEncoderHandle)
	MicrosoftSecret(container, id, secret).CodecEncodeSelf(encoder)
	fmt.Fprintf(w, "\n")
	return nil
}

func labels(name string) map[string]string {
	return map[string]string{
		"app":   name,
		"suite": suite,
	}
}

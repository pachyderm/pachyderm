package assets

import (
	"encoding/json"
	"fmt"
	"io"
	"path"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/pachyderm/pachyderm/src/client"
	auth "github.com/pachyderm/pachyderm/src/server/auth/server"
	pfs "github.com/pachyderm/pachyderm/src/server/pfs/server"
	"github.com/pachyderm/pachyderm/src/server/pps/server/githook"
	apps "k8s.io/api/apps/v1beta2"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	suite = "pachyderm"

	pachdImage = "pachyderm/pachd"
	// Using our own etcd image for now because there's a fix we need
	// that hasn't been released, and which has been manually applied
	// to the official v3.2.7 release.
	etcdImage      = "pachyderm/etcd:v3.2.7"
	grpcProxyImage = "pachyderm/grpc-proxy:0.4.2"
	dashName       = "dash"
	workerImage    = "pachyderm/worker"
	pauseImage     = "gcr.io/google_containers/pause-amd64:3.0"
	dashImage      = "pachyderm/dash"

	serviceAccountName      = "pachyderm"
	etcdHeadlessServiceName = "etcd-headless"
	etcdName                = "etcd"
	etcdVolumeName          = "etcd-volume"
	etcdVolumeClaimName     = "etcd-storage"
	etcdStorageClassName    = "etcd-storage-class"
	grpcProxyName           = "grpc-proxy"
	pachdName               = "pachd"

	trueVal = true
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
	Registry    string

	// NoGuaranteed will not generate assets that have both resource limits and
	// resource requests set which causes kubernetes to give the pods
	// guaranteed QoS. Guaranteed QoS generally leads to more stable clusters
	// but on smaller test clusters such as those run on minikube it doesn't
	// help much and may cause more instability than it prevents.
	NoGuaranteed bool

	// DisableAuthentication stops Pachyderm's authentication service
	// from talking to GitHub, for testing. Instead users can authenticate
	// simply by providing a username.
	DisableAuthentication bool

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

	// IAMRole is the IAM role that the Pachyderm deployment should
	// assume when talking to AWS services.
	IAMRole string

	// ImagePullSecret specifies an image pull secret that gets attached to the
	// various deployments so that their images can be pulled from a private
	// registry.
	ImagePullSecret string
}

// replicas lets us create a pointer to a non-zero int32 in-line. This is
// helpful because the Replicas field of many specs expectes an *int32
func replicas(r int32) *int32 {
	return &r
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
func ServiceAccount() *v1.ServiceAccount {
	return &v1.ServiceAccount{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ServiceAccount",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:   serviceAccountName,
			Labels: labels(""),
		},
	}
}

// GetSecretVolumeAndMount returns a properly configured Volume and
// VolumeMount object given a backend.  The backend needs to be one of the
// constants defined in pfs/server.
func GetSecretVolumeAndMount(backend string) (v1.Volume, v1.VolumeMount) {
	return v1.Volume{
			Name: client.StorageSecretName,
			VolumeSource: v1.VolumeSource{
				Secret: &v1.SecretVolumeSource{
					SecretName: client.StorageSecretName,
				},
			},
		}, v1.VolumeMount{
			Name:      client.StorageSecretName,
			MountPath: "/" + client.StorageSecretName,
		}
}

func versionedPachdImage(opts *AssetOpts) string {
	if opts.Version != "" {
		return fmt.Sprintf("%s:%s", pachdImage, opts.Version)
	}
	return pachdImage
}

func versionedWorkerImage(opts *AssetOpts) string {
	if opts.Version != "" {
		return fmt.Sprintf("%s:%s", workerImage, opts.Version)
	}
	return workerImage
}

func imagePullSecrets(opts *AssetOpts) []v1.LocalObjectReference {
	var result []v1.LocalObjectReference
	if opts.ImagePullSecret != "" {
		result = append(result, v1.LocalObjectReference{Name: opts.ImagePullSecret})
	}
	return result
}

// PachdDeployment returns a pachd k8s Deployment.
func PachdDeployment(opts *AssetOpts, objectStoreBackend backend, hostPath string) *apps.Deployment {
	mem := resource.MustParse(opts.BlockCacheSize)
	mem.Add(resource.MustParse(opts.PachdNonCacheMemRequest))
	cpu := resource.MustParse(opts.PachdCPURequest)
	image := AddRegistry(opts.Registry, versionedPachdImage(opts))
	volumes := []v1.Volume{
		{
			Name: "pach-disk",
		},
	}
	volumeMounts := []v1.VolumeMount{
		{
			Name:      "pach-disk",
			MountPath: "/pach",
		},
	}

	// Set up storage options
	var backendEnvVar string
	var storageHostPath string
	switch objectStoreBackend {
	case localBackend:
		storageHostPath = filepath.Join(hostPath, "pachd")
		volumes[0].HostPath = &v1.HostPathVolumeSource{
			Path: storageHostPath,
		}
		backendEnvVar = pfs.LocalBackendEnvVar
	case minioBackend:
		backendEnvVar = pfs.MinioBackendEnvVar
	case amazonBackend:
		backendEnvVar = pfs.AmazonBackendEnvVar
	case googleBackend:
		backendEnvVar = pfs.GoogleBackendEnvVar
	case microsoftBackend:
		backendEnvVar = pfs.MicrosoftBackendEnvVar
	}
	volume, mount := GetSecretVolumeAndMount(backendEnvVar)
	volumes = append(volumes, volume)
	volumeMounts = append(volumeMounts, mount)
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
	return &apps.Deployment{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Deployment",
			APIVersion: "apps/v1beta2",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:   pachdName,
			Labels: labels(pachdName),
		},
		Spec: apps.DeploymentSpec{
			Replicas: replicas(1),
			Selector: &metav1.LabelSelector{
				MatchLabels: labels(pachdName),
			},
			Template: v1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Name:   pachdName,
					Labels: labels(pachdName),
					Annotations: map[string]string{
						"iam.amazonaws.com/role": opts.IAMRole,
					},
				},
				Spec: v1.PodSpec{
					Containers: []v1.Container{
						{
							Name:  pachdName,
							Image: image,
							Env: []v1.EnvVar{
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
									ValueFrom: &v1.EnvVarSource{
										FieldRef: &v1.ObjectFieldSelector{
											APIVersion: "v1",
											FieldPath:  "metadata.namespace",
										},
									},
								},
								{
									Name:  "WORKER_IMAGE",
									Value: AddRegistry(opts.Registry, versionedWorkerImage(opts)),
								},
								{
									Name:  "IMAGE_PULL_SECRET",
									Value: opts.ImagePullSecret,
								},
								{
									Name:  "WORKER_SIDECAR_IMAGE",
									Value: image,
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
								{
									Name:  "IAM_ROLE",
									Value: opts.IAMRole,
								},
								{
									Name:  auth.DisableAuthenticationEnvVar,
									Value: strconv.FormatBool(opts.DisableAuthentication),
								},
							},
							Ports: []v1.ContainerPort{
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
									ContainerPort: pfs.HTTPPort,
									Protocol:      "TCP",
									Name:          "api-http-port",
								},
								{
									ContainerPort: githook.GitHookPort,
									Protocol:      "TCP",
									Name:          "api-git-port",
								},
							},
							VolumeMounts: volumeMounts,
							SecurityContext: &v1.SecurityContext{
								Privileged: &trueVal, // god is this dumb
							},
							ImagePullPolicy: "IfNotPresent",
							Resources:       resourceRequirements,
						},
					},
					ServiceAccountName: serviceAccountName,
					Volumes:            volumes,
					ImagePullSecrets:   imagePullSecrets(opts),
				},
			},
		},
	}
}

// PachdService returns a pachd service.
func PachdService() *v1.Service {
	return &v1.Service{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Service",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
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
					Port:     pfs.HTTPPort,
					Name:     "api-http-port",
					NodePort: 30000 + pfs.HTTPPort,
				},
				{
					Port:     githook.GitHookPort,
					Name:     "api-git-port",
					NodePort: 30000 + githook.GitHookPort,
				},
			},
		},
	}
}

// EtcdDeployment returns an etcd k8s Deployment.
func EtcdDeployment(opts *AssetOpts, hostPath string) *apps.Deployment {
	cpu := resource.MustParse(opts.EtcdCPURequest)
	mem := resource.MustParse(opts.EtcdMemRequest)
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
		volumes = []v1.Volume{
			{
				Name: "etcd-storage",
				VolumeSource: v1.VolumeSource{
					HostPath: &v1.HostPathVolumeSource{
						Path: filepath.Join(hostPath, "etcd"),
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
	return &apps.Deployment{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Deployment",
			APIVersion: "apps/v1beta2",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:   etcdName,
			Labels: labels(etcdName),
		},
		Spec: apps.DeploymentSpec{
			Replicas: replicas(1),
			Selector: &metav1.LabelSelector{
				MatchLabels: labels(etcdName),
			},
			Template: v1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Name:   etcdName,
					Labels: labels(etcdName),
				},
				Spec: v1.PodSpec{
					Containers: []v1.Container{
						{
							Name:  etcdName,
							Image: AddRegistry(opts.Registry, etcdImage),
							//TODO figure out how to get a cluster of these to talk to each other
							Command: []string{
								"/usr/local/bin/etcd",
								"--listen-client-urls=http://0.0.0.0:2379",
								"--advertise-client-urls=http://0.0.0.0:2379",
								"--data-dir=/var/data/etcd",
								"--auto-compaction-retention=1",
							},
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
	hostPath string, name string, size int) (*v1.PersistentVolume, error) {
	spec := &v1.PersistentVolume{
		TypeMeta: metav1.TypeMeta{
			Kind:       "PersistentVolume",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:   etcdVolumeName,
			Labels: labels(etcdName),
		},
		Spec: v1.PersistentVolumeSpec{
			Capacity: map[v1.ResourceName]resource.Quantity{
				"storage": resource.MustParse(fmt.Sprintf("%vGi", size)),
			},
			AccessModes:                   []v1.PersistentVolumeAccessMode{v1.ReadWriteOnce},
			PersistentVolumeReclaimPolicy: v1.PersistentVolumeReclaimRetain,
		},
	}

	switch persistentDiskBackend {
	case amazonBackend:
		spec.Spec.PersistentVolumeSource = v1.PersistentVolumeSource{
			AWSElasticBlockStore: &v1.AWSElasticBlockStoreVolumeSource{
				FSType:   "ext4",
				VolumeID: name,
			},
		}
	case googleBackend:
		spec.Spec.PersistentVolumeSource = v1.PersistentVolumeSource{
			GCEPersistentDisk: &v1.GCEPersistentDiskVolumeSource{
				FSType: "ext4",
				PDName: name,
			},
		}
	case microsoftBackend:
		dataDiskURI := name
		split := strings.Split(name, "/")
		diskName := split[len(split)-1]

		spec.Spec.PersistentVolumeSource = v1.PersistentVolumeSource{
			AzureDisk: &v1.AzureDiskVolumeSource{
				DiskName:    diskName,
				DataDiskURI: dataDiskURI,
			},
		}
	case minioBackend:
		fallthrough
	case localBackend:
		spec.Spec.PersistentVolumeSource = v1.PersistentVolumeSource{
			HostPath: &v1.HostPathVolumeSource{
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
func EtcdVolumeClaim(size int) *v1.PersistentVolumeClaim {
	return &v1.PersistentVolumeClaim{
		TypeMeta: metav1.TypeMeta{
			Kind:       "PersistentVolumeClaim",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:   etcdVolumeClaimName,
			Labels: labels(etcdName),
		},
		Spec: v1.PersistentVolumeClaimSpec{
			Resources: v1.ResourceRequirements{
				Requests: map[v1.ResourceName]resource.Quantity{
					"storage": resource.MustParse(fmt.Sprintf("%vGi", size)),
				},
			},
			AccessModes: []v1.PersistentVolumeAccessMode{v1.ReadWriteOnce},
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
		TypeMeta: metav1.TypeMeta{
			Kind:       "Service",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
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
		TypeMeta: metav1.TypeMeta{
			Kind:       "Service",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
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
	var imagePullSecrets []map[string]string
	if opts.ImagePullSecret != "" {
		imagePullSecrets = append(imagePullSecrets, map[string]string{"name": opts.ImagePullSecret})
	}
	// As of March 17, 2017, the Kubernetes client does not include structs for
	// Stateful Set, so we generate the kubernetes manifest using raw json.
	// TODO(msteffen): we're now upgrading our kubernetes client, so we should be
	// abe to rewrite this spec using k8s client structs
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
							"image":   AddRegistry(opts.Registry, etcdImage),
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
			"imagePullSecrets":     imagePullSecrets,
		},
	}
}

// DashDeployment creates a Deployment for the pachyderm dashboard.
func DashDeployment(opts *AssetOpts) *apps.Deployment {
	return &apps.Deployment{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Deployment",
			APIVersion: "apps/v1beta2",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:   dashName,
			Labels: labels(dashName),
		},
		Spec: apps.DeploymentSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: labels(dashName),
			},
			Template: v1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Name:   dashName,
					Labels: labels(dashName),
				},
				Spec: v1.PodSpec{
					Containers: []v1.Container{
						{
							Name:  dashName,
							Image: AddRegistry(opts.Registry, opts.DashImage),
							Ports: []v1.ContainerPort{
								{
									ContainerPort: 8080,
									Name:          "dash-http",
								},
							},
							ImagePullPolicy: "IfNotPresent",
						},
						{
							Name:  grpcProxyName,
							Image: AddRegistry(opts.Registry, grpcProxyImage),
							Ports: []v1.ContainerPort{
								{
									ContainerPort: 8081,
									Name:          "grpc-proxy-http",
								},
							},
							ImagePullPolicy: "IfNotPresent",
						},
					},
					ImagePullSecrets: imagePullSecrets(opts),
				},
			},
		},
	}
}

// DashService creates a Service for the pachyderm dashboard.
func DashService() *v1.Service {
	return &v1.Service{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Service",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
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
func MinioSecret(bucket string, id string, secret string, endpoint string, secure bool) map[string][]byte {
	secureV := "0"
	if secure {
		secureV = "1"
	}
	return map[string][]byte{
		"minio-bucket":   []byte(bucket),
		"minio-id":       []byte(id),
		"minio-secret":   []byte(secret),
		"minio-endpoint": []byte(endpoint),
		"minio-secure":   []byte(secureV),
	}
}

// WriteSecret writes a JSON-encoded k8s secret to the given writer.
// The secret uses the given map as data.
func WriteSecret(w io.Writer, data map[string][]byte) {
	secret := &v1.Secret{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Secret",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:   client.StorageSecretName,
			Labels: labels(client.StorageSecretName),
		},
		Data: data,
	}
	encoder := json.NewEncoder(w)
	encoder.Encode(secret)
	fmt.Fprintf(w, "\n")
}

// LocalSecret creates an empty secret.
func LocalSecret() map[string][]byte {
	return nil
}

// AmazonSecret creates an amazon secret with the following parameters:
//   bucket - S3 bucket name
//   id     - AWS access key id
//   secret - AWS secret access key
//   token  - AWS access token
//   region - AWS region
func AmazonSecret(bucket string, distribution string, id string, secret string, token string, region string) map[string][]byte {
	return map[string][]byte{
		"amazon-bucket":       []byte(bucket),
		"amazon-distribution": []byte(distribution),
		"amazon-id":           []byte(id),
		"amazon-secret":       []byte(secret),
		"amazon-token":        []byte(token),
		"amazon-region":       []byte(region),
	}
}

// GoogleSecret creates a google secret with a bucket name.
func GoogleSecret(bucket string) map[string][]byte {
	return map[string][]byte{
		"google-bucket": []byte(bucket),
	}
}

// MicrosoftSecret creates a microsoft secret with following parameters:
//   container - Azure blob container
//   id    	   - Azure storage account name
//   secret    - Azure storage account key
func MicrosoftSecret(container string, id string, secret string) map[string][]byte {
	return map[string][]byte{
		"microsoft-container": []byte(container),
		"microsoft-id":        []byte(id),
		"microsoft-secret":    []byte(secret),
	}
}

// WriteDashboardAssets writes the k8s config for deploying the Pachyderm
// dashboard to 'w'
func WriteDashboardAssets(w io.Writer, opts *AssetOpts) {
	encoder := json.NewEncoder(w)
	encoder.Encode(DashService())
	fmt.Fprintf(w, "\n")
	encoder.Encode(DashDeployment(opts))
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
	encoder := json.NewEncoder(w)

	encoder.Encode(ServiceAccount())
	fmt.Fprintf(w, "\n")

	if opts.EtcdNodes > 0 && opts.EtcdVolume != "" {
		return fmt.Errorf("only one of --dynamic-etcd-nodes and --static-etcd-volume should be given, but not both")
	}

	// In the dynamic route, we create a storage class which dynamically
	// provisions volumes, and run etcd as a statful set.
	// In the static route, we create a single volume, a single volume
	// claim, and run etcd as a replication controller with a single node.
	if objectStoreBackend == localBackend {
		encoder.Encode(EtcdDeployment(opts, hostPath))
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
		encoder.Encode(EtcdHeadlessService())
		fmt.Fprintf(w, "\n")
		encoder.Encode(EtcdStatefulSet(opts, persistentDiskBackend, volumeSize))
		fmt.Fprintf(w, "\n")
	} else if opts.EtcdVolume != "" || persistentDiskBackend == localBackend {
		volume, err := EtcdVolume(persistentDiskBackend, opts, hostPath, opts.EtcdVolume, volumeSize)
		if err != nil {
			return err
		}
		encoder.Encode(volume)
		fmt.Fprintf(w, "\n")
		encoder.Encode(EtcdVolumeClaim(volumeSize))
		fmt.Fprintf(w, "\n")
		encoder.Encode(EtcdDeployment(opts, ""))
		fmt.Fprintf(w, "\n")
	} else {
		return fmt.Errorf("unless deploying locally, either --dynamic-etcd-nodes or --static-etcd-volume needs to be provided")
	}
	encoder.Encode(EtcdNodePortService(objectStoreBackend == localBackend))
	fmt.Fprintf(w, "\n")

	encoder.Encode(PachdService())
	fmt.Fprintf(w, "\n")
	encoder.Encode(PachdDeployment(opts, objectStoreBackend, hostPath))
	fmt.Fprintf(w, "\n")
	if opts.EnableDash {
		WriteDashboardAssets(w, opts)
	}
	return nil
}

// WriteLocalAssets writes assets to a local backend.
func WriteLocalAssets(w io.Writer, opts *AssetOpts, hostPath string) error {
	if err := WriteAssets(w, opts, localBackend, localBackend, 1 /* = volume size (gb) */, hostPath); err != nil {
		return err
	}
	WriteSecret(w, LocalSecret())
	return nil
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
		WriteSecret(w, MinioSecret(args[2], args[3], args[4], args[5], secure))
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
	WriteSecret(w, AmazonSecret(bucket, distribution, id, secret, token, region))
	return nil
}

// WriteGoogleAssets writes assets to a google backend.
func WriteGoogleAssets(w io.Writer, opts *AssetOpts, bucket string, volumeSize int) error {
	if err := WriteAssets(w, opts, googleBackend, googleBackend, volumeSize, ""); err != nil {
		return err
	}
	WriteSecret(w, GoogleSecret(bucket))
	return nil
}

// WriteMicrosoftAssets writes assets to a microsoft backend
func WriteMicrosoftAssets(w io.Writer, opts *AssetOpts, container string, id string, secret string, volumeSize int) error {
	if err := WriteAssets(w, opts, microsoftBackend, microsoftBackend, volumeSize, ""); err != nil {
		return err
	}
	WriteSecret(w, MicrosoftSecret(container, id, secret))
	return nil
}

// Images returns a list of all the images that are used by a pachyderm deployment.
func Images(opts *AssetOpts) []string {
	return []string{
		versionedWorkerImage(opts),
		etcdImage,
		grpcProxyImage,
		pauseImage,
		versionedPachdImage(opts),
		opts.DashImage,
	}
}

func labels(name string) map[string]string {
	return map[string]string{
		"app":   name,
		"suite": suite,
	}
}

// AddRegistry switchs the registry that an image is targetting.
func AddRegistry(registry string, imageName string) string {
	parts := strings.Split(imageName, "/")
	if len(parts) == 3 {
		parts = parts[1:]
	}
	if registry != "" {
		return path.Join(registry, parts[0], parts[1])
	}
	return path.Join(parts[0], parts[1])
}

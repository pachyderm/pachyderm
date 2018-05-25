package assets

import (
	"fmt"
	"path"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/pachyderm/pachyderm/src/client"
	auth "github.com/pachyderm/pachyderm/src/server/auth/server"
	"github.com/pachyderm/pachyderm/src/server/http"
	pfs "github.com/pachyderm/pachyderm/src/server/pfs/server"
	"github.com/pachyderm/pachyderm/src/server/pps/server/githook"
	apps "k8s.io/api/apps/v1beta1"
	"k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

var (
	suite = "pachyderm"

	pachdImage = "pachyderm/pachd"
	// Using our own etcd image for now because there's a fix we need
	// that hasn't been released, and which has been manually applied
	// to the official v3.2.7 release.
	etcdImage      = "quay.io/coreos/etcd:v3.3.5"
	grpcProxyImage = "pachyderm/grpc-proxy:0.4.2"
	dashName       = "dash"
	workerImage    = "pachyderm/worker"
	pauseImage     = "gcr.io/google_containers/pause-amd64:3.0"
	dashImage      = "pachyderm/dash"

	// ServiceAccountName is the name of Pachyderm's service account.
	// It's public because it's needed by pps.APIServer to create the RCs for
	// workers.
	ServiceAccountName      = "pachyderm"
	etcdHeadlessServiceName = "etcd-headless"
	etcdName                = "etcd"
	etcdVolumeName          = "etcd-volume"
	etcdVolumeClaimName     = "etcd-storage"
	// The storage class name to use when creating a new StorageClass for etcd.
	defaultEtcdStorageClassName = "etcd-storage-class"
	grpcProxyName               = "grpc-proxy"
	pachdName                   = "pachd"
	// PrometheusPort hosts the prometheus stats for scraping
	PrometheusPort = 9091

	// Role & binding names, used for Roles or ClusterRoles and their associated
	// bindings.
	roleName        = "pachyderm"
	roleBindingName = "pachyderm"
	// Policy rules to use for Pachyderm's Role or ClusterRole.
	rolePolicyRules = []rbacv1.PolicyRule{{
		APIGroups: []string{""},
		Verbs:     []string{"get", "list", "watch"},
		Resources: []string{"nodes", "pods", "pods/log", "endpoints"},
	}, {
		APIGroups: []string{""},
		Verbs:     []string{"get", "list", "watch", "create", "update", "delete"},
		Resources: []string{"replicationcontrollers", "services"},
	}, {
		APIGroups:     []string{""},
		Verbs:         []string{"get", "list", "watch", "create", "update", "delete"},
		Resources:     []string{"secrets"},
		ResourceNames: []string{client.StorageSecretName},
	}}
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
	DashOnly    bool
	NoDash      bool
	DashImage   string
	Registry    string
	EtcdPrefix  string

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

	// EtcdStorageClassName is the name of an existing StorageClass to use when
	// creating a StatefulSet for dynamic etcd storage. If unset, a new
	// StorageClass will be created for the StatefulSet.
	EtcdStorageClassName string

	// IAM role that the Pachyderm deployment should assume when talking to AWS
	// services (if using kube2iam + metadata service + IAM role to delegate
	// permissions to pachd via its instance).
	// This is in AssetOpts rather than AmazonCreds because it must be passed
	// as an annotation on the pachd pod rather than as a k8s secret
	IAMRole string

	// ImagePullSecret specifies an image pull secret that gets attached to the
	// various deployments so that their images can be pulled from a private
	// registry.
	ImagePullSecret string

	// NoRBAC, if true, will disable creation of RBAC assets.
	NoRBAC bool

	// LocalRoles, if true, uses Role and RoleBinding instead of ClusterRole and
	// ClusterRoleBinding.
	LocalRoles bool

	// Namespace is the kubernetes namespace to deploy to.
	Namespace string

	// NoExposeDockerSocket if true prevents pipelines from accessing the docker socket.
	NoExposeDockerSocket bool
}

// Encoder is the interface for writing out assets. This is assumed to wrap an output writer.
type Encoder interface {
	// Encodes the given struct to the wrapped output stream. This also will write out a separator
	// value, suitable for differentiating multiple objects in the stream.
	Encode(interface{}) (err error)
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
func ServiceAccount(opts *AssetOpts) *v1.ServiceAccount {
	return &v1.ServiceAccount{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ServiceAccount",
			APIVersion: "v1",
		},
		ObjectMeta: objectMeta(ServiceAccountName, labels(""), nil, opts.Namespace),
	}
}

// ClusterRole returns a ClusterRole that should be bound to the Pachyderm service account.
func ClusterRole(opts *AssetOpts) *rbacv1.ClusterRole {
	return &rbacv1.ClusterRole{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ClusterRole",
			APIVersion: "rbac.authorization.k8s.io/v1",
		},
		ObjectMeta: objectMeta(roleName, labels(""), nil, opts.Namespace),
		Rules:      rolePolicyRules,
	}
}

// ClusterRoleBinding returns a ClusterRoleBinding that binds Pachyderm's
// ClusterRole to its ServiceAccount.
func ClusterRoleBinding(opts *AssetOpts) *rbacv1.ClusterRoleBinding {
	return &rbacv1.ClusterRoleBinding{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ClusterRoleBinding",
			APIVersion: "rbac.authorization.k8s.io/v1",
		},
		ObjectMeta: objectMeta(roleBindingName, labels(""), nil, opts.Namespace),
		Subjects: []rbacv1.Subject{{
			Kind:      "ServiceAccount",
			Name:      ServiceAccountName,
			Namespace: opts.Namespace,
		}},
		RoleRef: rbacv1.RoleRef{
			Kind: "ClusterRole",
			Name: roleName,
		},
	}
}

// Role returns a Role that should be bound to the Pachyderm service account.
func Role(opts *AssetOpts) *rbacv1.Role {
	return &rbacv1.Role{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Role",
			APIVersion: "rbac.authorization.k8s.io/v1",
		},
		ObjectMeta: objectMeta(roleName, labels(""), nil, opts.Namespace),
		Rules:      rolePolicyRules,
	}
}

// RoleBinding returns a RoleBinding that binds Pachyderm's Role to its
// ServiceAccount.
func RoleBinding(opts *AssetOpts) *rbacv1.RoleBinding {
	return &rbacv1.RoleBinding{
		TypeMeta: metav1.TypeMeta{
			Kind:       "RoleBinding",
			APIVersion: "rbac.authorization.k8s.io/v1",
		},
		ObjectMeta: objectMeta(roleBindingName, labels(""), nil, opts.Namespace),
		Subjects: []rbacv1.Subject{{
			Kind:      "ServiceAccount",
			Name:      ServiceAccountName,
			Namespace: opts.Namespace,
		}},
		RoleRef: rbacv1.RoleRef{
			Kind: "Role",
			Name: roleName,
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
			APIVersion: "apps/v1beta1",
		},
		ObjectMeta: objectMeta(pachdName, labels(pachdName), nil, opts.Namespace),
		Spec: apps.DeploymentSpec{
			Replicas: replicas(1),
			Selector: &metav1.LabelSelector{
				MatchLabels: labels(pachdName),
			},
			Template: v1.PodTemplateSpec{
				ObjectMeta: objectMeta(pachdName, labels(pachdName),
					map[string]string{"iam.amazonaws.com/role": opts.IAMRole}, opts.Namespace),
				Spec: v1.PodSpec{
					Containers: []v1.Container{
						{
							Name:  pachdName,
							Image: image,
							Env: []v1.EnvVar{
								{Name: "PACH_ROOT", Value: "/pach"},
								{Name: "ETCD_PREFIX", Value: opts.EtcdPrefix},
								{Name: "NUM_SHARDS", Value: fmt.Sprintf("%d", opts.PachdShards)},
								{Name: "STORAGE_BACKEND", Value: backendEnvVar},
								{Name: "STORAGE_HOST_PATH", Value: storageHostPath},
								{Name: "WORKER_IMAGE", Value: AddRegistry(opts.Registry, versionedWorkerImage(opts))},
								{Name: "IMAGE_PULL_SECRET", Value: opts.ImagePullSecret},
								{Name: "WORKER_SIDECAR_IMAGE", Value: image},
								{Name: "WORKER_IMAGE_PULL_POLICY", Value: "IfNotPresent"},
								{Name: "PACHD_VERSION", Value: opts.Version},
								{Name: "METRICS", Value: strconv.FormatBool(opts.Metrics)},
								{Name: "LOG_LEVEL", Value: opts.LogLevel},
								{Name: "BLOCK_CACHE_BYTES", Value: opts.BlockCacheSize},
								{Name: "IAM_ROLE", Value: opts.IAMRole},
								{Name: "NO_EXPOSE_DOCKER_SOCKET", Value: strconv.FormatBool(opts.NoExposeDockerSocket)},
								{Name: auth.DisableAuthenticationEnvVar, Value: strconv.FormatBool(opts.DisableAuthentication)},
								{
									Name: "PACHD_POD_NAMESPACE",
									ValueFrom: &v1.EnvVarSource{
										FieldRef: &v1.ObjectFieldSelector{
											APIVersion: "v1",
											FieldPath:  "metadata.namespace",
										},
									},
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
									ContainerPort: http.HTTPPort,
									Protocol:      "TCP",
									Name:          "api-http-port",
								},
								{
									ContainerPort: githook.GitHookPort,
									Protocol:      "TCP",
									Name:          "api-git-port",
								},
							},
							VolumeMounts:    volumeMounts,
							ImagePullPolicy: "IfNotPresent",
							Resources:       resourceRequirements,
							ReadinessProbe: &v1.Probe{
								Handler: v1.Handler{
									Exec: &v1.ExecAction{
										Command: []string{"/pachd", "--readiness"},
									},
								},
							},
						},
					},
					ServiceAccountName: ServiceAccountName,
					Volumes:            volumes,
					ImagePullSecrets:   imagePullSecrets(opts),
				},
			},
		},
	}
}

// PachdService returns a pachd service.
func PachdService(opts *AssetOpts) *v1.Service {
	prometheusAnnotations := map[string]string{
		"prometheus.io/scrape": "true",
		"prometheus.io/port":   string(PrometheusPort),
	}
	return &v1.Service{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Service",
			APIVersion: "v1",
		},
		ObjectMeta: objectMeta(pachdName, labels(pachdName), prometheusAnnotations, opts.Namespace),
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
					Port:     http.HTTPPort,
					Name:     "api-http-port",
					NodePort: 30000 + http.HTTPPort,
				},
				{
					Port:     githook.GitHookPort,
					Name:     "api-git-port",
					NodePort: githook.NodePort(),
				},
			},
		},
	}
}

// GithookService returns a k8s service that exposes a public IP
func GithookService(namespace string) *v1.Service {
	name := "githook"
	return &v1.Service{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Service",
			APIVersion: "v1",
		},
		ObjectMeta: objectMeta(name, labels(name), nil, namespace),
		Spec: v1.ServiceSpec{
			Type: v1.ServiceTypeLoadBalancer,
			Selector: map[string]string{
				"app": pachdName,
			},
			Ports: []v1.ServicePort{
				{
					TargetPort: intstr.FromInt(githook.GitHookPort),
					Name:       "api-git-port",
					Port:       githook.ExternalPort(),
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
	// Don't want to strip the registry out of etcdImage since it's from quay
	// not docker hub.
	image := etcdImage
	if opts.Registry != "" {
		image = AddRegistry(opts.Registry, etcdImage)
	}
	return &apps.Deployment{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Deployment",
			APIVersion: "apps/v1beta1",
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
							Command: []string{
								"/usr/local/bin/etcd",
								"--listen-client-urls=http://0.0.0.0:2379",
								"--advertise-client-urls=http://0.0.0.0:2379",
								"--data-dir=/var/data/etcd",
								"--auto-compaction-retention=1",
								"--max-txn-ops=5000",
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
func EtcdStorageClass(opts *AssetOpts, backend backend) (interface{}, error) {
	sc := map[string]interface{}{
		"apiVersion": "storage.k8s.io/v1beta1",
		"kind":       "StorageClass",
		"metadata": map[string]interface{}{
			"name":      defaultEtcdStorageClassName,
			"labels":    labels(etcdName),
			"namespace": opts.Namespace,
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
		ObjectMeta: objectMeta(etcdVolumeName, labels(etcdName), nil, opts.Namespace),
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
// unnecessary (the stateful set controller will create PVCs automatically).
func EtcdVolumeClaim(size int, opts *AssetOpts) *v1.PersistentVolumeClaim {
	return &v1.PersistentVolumeClaim{
		TypeMeta: metav1.TypeMeta{
			Kind:       "PersistentVolumeClaim",
			APIVersion: "v1",
		},
		ObjectMeta: objectMeta(etcdVolumeClaimName, labels(etcdName), nil, opts.Namespace),
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

// EtcdHeadlessService returns a headless etcd service, which is only for DNS
// resolution.
func EtcdHeadlessService(opts *AssetOpts) *v1.Service {
	return &v1.Service{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Service",
			APIVersion: "v1",
		},
		ObjectMeta: objectMeta(etcdHeadlessServiceName, labels(etcdName), nil, opts.Namespace),
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
		"--max-txn-ops=5000",
	}
	for i, str := range etcdCmd {
		etcdCmd[i] = fmt.Sprintf("\"%s\"", str) // quote all arguments, for shell
	}

	var pvcTemplates []interface{}
	switch backend {
	case googleBackend, amazonBackend:
		storageClassName := opts.EtcdStorageClassName
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
		"apiVersion": "apps/v1beta1",
		"kind":       "StatefulSet",
		"metadata": map[string]interface{}{
			"name":      etcdName,
			"labels":    labels(etcdName),
			"namespace": opts.Namespace,
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

// DashDeployment creates a Deployment for the pachyderm dashboard.
func DashDeployment(opts *AssetOpts) *apps.Deployment {
	return &apps.Deployment{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Deployment",
			APIVersion: "apps/v1beta1",
		},
		ObjectMeta: objectMeta(dashName, labels(dashName), nil, opts.Namespace),
		Spec: apps.DeploymentSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: labels(dashName),
			},
			Template: v1.PodTemplateSpec{
				ObjectMeta: objectMeta(dashName, labels(dashName), nil, opts.Namespace),
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
func DashService(opts *AssetOpts) *v1.Service {
	return &v1.Service{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Service",
			APIVersion: "v1",
		},
		ObjectMeta: objectMeta(dashName, labels(dashName), nil, opts.Namespace),
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
//   isS3V2 - Set to true if client follows S3V2
func MinioSecret(bucket string, id string, secret string, endpoint string, secure, isS3V2 bool) map[string][]byte {
	secureV := "0"
	if secure {
		secureV = "1"
	}
	s3V2 := "0"
	if isS3V2 {
		s3V2 = "1"
	}
	return map[string][]byte{
		"minio-bucket":    []byte(bucket),
		"minio-id":        []byte(id),
		"minio-secret":    []byte(secret),
		"minio-endpoint":  []byte(endpoint),
		"minio-secure":    []byte(secureV),
		"minio-signature": []byte(s3V2),
	}
}

// WriteSecret writes a JSON-encoded k8s secret to the given writer.
// The secret uses the given map as data.
func WriteSecret(encoder Encoder, data map[string][]byte, opts *AssetOpts) error {
	if opts.DashOnly {
		return nil
	}
	secret := &v1.Secret{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Secret",
			APIVersion: "v1",
		},
		ObjectMeta: objectMeta(client.StorageSecretName, labels(client.StorageSecretName), nil, opts.Namespace),
		Data:       data,
	}
	return encoder.Encode(secret)
}

// LocalSecret creates an empty secret.
func LocalSecret() map[string][]byte {
	return nil
}

// AmazonSecret creates an amazon secret with the following parameters:
//   region       - AWS region
//   bucket       - S3 bucket name
//   id           - AWS access key id
//   secret       - AWS secret access key
//   token        - AWS access token
//   distribution - cloudfront distribution
func AmazonSecret(region, bucket, id, secret, token, distribution string) map[string][]byte {
	return map[string][]byte{
		"amazon-region":       []byte(region),
		"amazon-bucket":       []byte(bucket),
		"amazon-id":           []byte(id),
		"amazon-secret":       []byte(secret),
		"amazon-token":        []byte(token),
		"amazon-distribution": []byte(distribution),
	}
}

// AmazonVaultSecret creates an amazon secret with the following parameters:
//   region       - AWS region
//   bucket       - S3 bucket name
//   vaultAddress - address/hostport of vault
//   vaultRole    - pachd's role in vault
//   vaultToken   - pachd's vault token
//   distribution - cloudfront distribution
func AmazonVaultSecret(region, bucket, vaultAddress, vaultRole, vaultToken, distribution string) map[string][]byte {
	return map[string][]byte{
		"amazon-region":       []byte(region),
		"amazon-bucket":       []byte(bucket),
		"amazon-vault-addr":   []byte(vaultAddress),
		"amazon-vault-role":   []byte(vaultRole),
		"amazon-vault-token":  []byte(vaultToken),
		"amazon-distribution": []byte(distribution),
	}
}

// AmazonIAMRoleSecret creates an amazon secret with the following parameters:
//   region       - AWS region
//   bucket       - S3 bucket name
//   distribution - cloudfront distribution
func AmazonIAMRoleSecret(region, bucket, distribution string) map[string][]byte {
	return map[string][]byte{
		"amazon-region":       []byte(region),
		"amazon-bucket":       []byte(bucket),
		"amazon-distribution": []byte(distribution),
	}
}

// GoogleSecret creates a google secret with a bucket name.
func GoogleSecret(bucket string, cred string) map[string][]byte {
	return map[string][]byte{
		"google-bucket": []byte(bucket),
		"google-cred":   []byte(cred),
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
// dashboard to 'encoder'
func WriteDashboardAssets(encoder Encoder, opts *AssetOpts) error {
	if err := encoder.Encode(DashService(opts)); err != nil {
		return err
	}
	return encoder.Encode(DashDeployment(opts))
}

// WriteAssets writes the assets to encoder.
func WriteAssets(encoder Encoder, opts *AssetOpts, objectStoreBackend backend,
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
		WriteDashboardAssets(encoder, opts)
		return nil
	}

	if err := encoder.Encode(ServiceAccount(opts)); err != nil {
		return err
	}
	if !opts.NoRBAC {
		if opts.LocalRoles {
			if err := encoder.Encode(Role(opts)); err != nil {
				return err
			}
			if err := encoder.Encode(RoleBinding(opts)); err != nil {
				return err
			}
		} else {
			if err := encoder.Encode(ClusterRole(opts)); err != nil {
				return err
			}
			if err := encoder.Encode(ClusterRoleBinding(opts)); err != nil {
				return err
			}
		}
	}

	if opts.EtcdNodes > 0 && opts.EtcdVolume != "" {
		return fmt.Errorf("only one of --dynamic-etcd-nodes and --static-etcd-volume should be given, but not both")
	}

	// In the dynamic route, we create a storage class which dynamically
	// provisions volumes, and run etcd as a statful set.
	// In the static route, we create a single volume, a single volume
	// claim, and run etcd as a replication controller with a single node.
	if objectStoreBackend == localBackend {
		if err := encoder.Encode(EtcdDeployment(opts, hostPath)); err != nil {
			return err
		}
	} else if opts.EtcdNodes > 0 {
		// Create a StorageClass, if the user didn't provide one.
		if opts.EtcdStorageClassName == "" {
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
	} else if opts.EtcdVolume != "" || persistentDiskBackend == localBackend {
		volume, err := EtcdVolume(persistentDiskBackend, opts, hostPath, opts.EtcdVolume, volumeSize)
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
		return fmt.Errorf("unless deploying locally, either --dynamic-etcd-nodes or --static-etcd-volume needs to be provided")
	}
	if err := encoder.Encode(EtcdNodePortService(objectStoreBackend == localBackend, opts)); err != nil {
		return err
	}

	if err := encoder.Encode(PachdService(opts)); err != nil {
		return err
	}
	if err := encoder.Encode(PachdDeployment(opts, objectStoreBackend, hostPath)); err != nil {
		return err
	}
	if !opts.NoDash {
		return WriteDashboardAssets(encoder, opts)
	}
	return nil
}

// WriteLocalAssets writes assets to a local backend.
func WriteLocalAssets(encoder Encoder, opts *AssetOpts, hostPath string) error {
	if err := WriteAssets(encoder, opts, localBackend, localBackend, 1 /* = volume size (gb) */, hostPath); err != nil {
		return err
	}
	WriteSecret(encoder, LocalSecret(), opts)
	return nil
}

// WriteCustomAssets writes assets to a custom combination of object-store and persistent disk.
func WriteCustomAssets(encoder Encoder, opts *AssetOpts, args []string, objectStoreBackend string,
	persistentDiskBackend string, secure, isS3V2 bool) error {
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
			if err := WriteAssets(encoder, opts, minioBackend, amazonBackend, volumeSize, ""); err != nil {
				return err
			}
		case "google":
			if err := WriteAssets(encoder, opts, minioBackend, googleBackend, volumeSize, ""); err != nil {
				return err
			}
		case "azure":
			if err := WriteAssets(encoder, opts, minioBackend, microsoftBackend, volumeSize, ""); err != nil {
				return err
			}
		default:
			return fmt.Errorf("Did not recognize the choice of persistent-disk")
		}
		return WriteSecret(encoder, MinioSecret(args[2], args[3], args[4], args[5], secure, isS3V2), opts)
	default:
		return fmt.Errorf("Did not recognize the choice of object-store")
	}
}

// AmazonCreds are options that are applicable specifically to Pachd's
// credentials in an AWS deployment
type AmazonCreds struct {
	// Direct credentials. Only applicable if Pachyderm is given its own permanent
	// AWS credentials
	ID     string // Access Key ID
	Secret string // Secret Access Key
	Token  string // Access token (if using temporary security credentials

	// Vault options (if getting AWS credentials from Vault)
	VaultAddress string // normally addresses come from env, but don't have vault service name
	VaultRole    string
	VaultToken   string
}

// WriteAmazonAssets writes assets to an amazon backend.
func WriteAmazonAssets(encoder Encoder, opts *AssetOpts, region string, bucket string, volumeSize int, creds *AmazonCreds, cloudfrontDistro string) error {
	if err := WriteAssets(encoder, opts, amazonBackend, amazonBackend, volumeSize, ""); err != nil {
		return err
	}
	var secret map[string][]byte
	if creds == nil {
		secret = AmazonIAMRoleSecret(region, bucket, cloudfrontDistro)
	} else if creds.ID != "" {
		secret = AmazonSecret(region, bucket, creds.ID, creds.Secret, creds.Token, cloudfrontDistro)
	} else if creds.VaultAddress != "" {
		secret = AmazonVaultSecret(region, bucket, creds.VaultAddress, creds.VaultRole, creds.VaultToken, cloudfrontDistro)
	}
	return WriteSecret(encoder, secret, opts)
}

// WriteGoogleAssets writes assets to a google backend.
func WriteGoogleAssets(encoder Encoder, opts *AssetOpts, bucket string, cred string, volumeSize int) error {
	if err := WriteAssets(encoder, opts, googleBackend, googleBackend, volumeSize, ""); err != nil {
		return err
	}
	return WriteSecret(encoder, GoogleSecret(bucket, cred), opts)
}

// WriteMicrosoftAssets writes assets to a microsoft backend
func WriteMicrosoftAssets(encoder Encoder, opts *AssetOpts, container string, id string, secret string, volumeSize int) error {
	if err := WriteAssets(encoder, opts, microsoftBackend, microsoftBackend, volumeSize, ""); err != nil {
		return err
	}
	return WriteSecret(encoder, MicrosoftSecret(container, id, secret), opts)
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

func objectMeta(name string, labels, annotations map[string]string, namespace string) metav1.ObjectMeta {
	return metav1.ObjectMeta{
		Name:        name,
		Labels:      labels,
		Annotations: annotations,
		Namespace:   namespace,
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

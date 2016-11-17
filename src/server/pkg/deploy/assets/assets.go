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
	"k8s.io/kubernetes/pkg/api/resource"
	"k8s.io/kubernetes/pkg/api/unversioned"
	api "k8s.io/kubernetes/pkg/api/v1"
	extensions "k8s.io/kubernetes/pkg/apis/extensions/v1beta1"
)

var (
	suite                  = "pachyderm"
	volumeSuite            = "pachyderm-pps-storage"
	pachdImage             = "pachyderm/pachd"
	jobShimImage           = "pachyderm/job-shim"
	etcdImage              = "gcr.io/google_containers/etcd:2.0.12"
	rethinkImage           = "rethinkdb:2.3.5"
	serviceAccountName     = "pachyderm"
	etcdName               = "etcd"
	pachdName              = "pachd"
	rethinkName            = "rethink"
	rethinkVolumeName      = "rethink-volume"
	rethinkVolumeClaimName = "rethink-volume-claim"
	amazonSecretName       = "amazon-secret"
	googleSecretName       = "google-secret"
	microsoftSecretName    = "microsoft-secret"
	initName               = "pachd-init"
	jobShimInitName        = "job-shim-init"
	trueVal                = true
	jsonEncoderHandle      = &codec.JsonHandle{
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
)

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

// PachdRc returns a pachd replication controller.
func PachdRc(shards uint64, backend backend, hostPath string, logLevel string, version string) *api.ReplicationController {
	image := pachdImage
	if version != "" {
		image += ":" + version
	}
	// we turn metrics on only if we have a static version this prevents dev
	// clusters from reporting metrics
	metrics := "true"
	if version == deploy.DevVersionTag {
		metrics = "false"
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
	readinessProbe := &api.Probe{
		Handler: api.Handler{
			Exec: &api.ExecAction{
				Command: []string{
					"./pachd",
					"--readiness-check",
				},
			},
		},
		InitialDelaySeconds: 15,
		TimeoutSeconds:      1,
	}
	var backendEnvVar string
	switch backend {
	case localBackend:
		volumes[0].HostPath = &api.HostPathVolumeSource{
			Path: filepath.Join(hostPath, "pachd"),
		}
	case amazonBackend:
		backendEnvVar = server.AmazonBackendEnvVar
		volumes = append(volumes, api.Volume{
			Name: amazonSecretName,
			VolumeSource: api.VolumeSource{
				Secret: &api.SecretVolumeSource{
					SecretName: amazonSecretName,
				},
			},
		})
		volumeMounts = append(volumeMounts, api.VolumeMount{
			Name:      amazonSecretName,
			MountPath: "/" + amazonSecretName,
		})
	case googleBackend:
		backendEnvVar = server.GoogleBackendEnvVar
		volumes = append(volumes, api.Volume{
			Name: googleSecretName,
			VolumeSource: api.VolumeSource{
				Secret: &api.SecretVolumeSource{
					SecretName: googleSecretName,
				},
			},
		})
		volumeMounts = append(volumeMounts, api.VolumeMount{
			Name:      googleSecretName,
			MountPath: "/" + googleSecretName,
		})
	case microsoftBackend:
		backendEnvVar = server.MicrosoftBackendEnvVar
		volumes = append(volumes, api.Volume{
			Name: microsoftSecretName,
			VolumeSource: api.VolumeSource{
				Secret: &api.SecretVolumeSource{
					SecretName: microsoftSecretName,
				},
			},
		})
		volumeMounts = append(volumeMounts, api.VolumeMount{
			Name:      microsoftSecretName,
			MountPath: "/" + microsoftSecretName,
		})
	}
	replicas := int32(1)
	return &api.ReplicationController{
		TypeMeta: unversioned.TypeMeta{
			Kind:       "ReplicationController",
			APIVersion: "v1",
		},
		ObjectMeta: api.ObjectMeta{
			Name:   pachdName,
			Labels: labels(pachdName),
		},
		Spec: api.ReplicationControllerSpec{
			Replicas: &replicas,
			Selector: map[string]string{
				"app": pachdName,
			},
			Template: &api.PodTemplateSpec{
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
									Value: strconv.FormatUint(shards, 10),
								},
								{
									Name:  "STORAGE_BACKEND",
									Value: backendEnvVar,
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
									Name:  "JOB_SHIM_IMAGE",
									Value: fmt.Sprintf("pachyderm/job-shim:%s", version),
								},
								{
									Name:  "JOB_IMAGE_PULL_POLICY",
									Value: "IfNotPresent",
								},
								{
									Name:  "PACHD_VERSION",
									Value: version,
								},
								{
									Name:  "METRICS",
									Value: metrics,
								},
								{
									Name:  "LOG_LEVEL",
									Value: logLevel,
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
							},
							VolumeMounts: volumeMounts,
							SecurityContext: &api.SecurityContext{
								Privileged: &trueVal, // god is this dumb
							},
							ReadinessProbe:  readinessProbe,
							ImagePullPolicy: "IfNotPresent",
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
func PachdService() *api.Service {
	return &api.Service{
		TypeMeta: unversioned.TypeMeta{
			Kind:       "Service",
			APIVersion: "v1",
		},
		ObjectMeta: api.ObjectMeta{
			Name:   pachdName,
			Labels: labels(pachdName),
		},
		Spec: api.ServiceSpec{
			Type: api.ServiceTypeNodePort,
			Selector: map[string]string{
				"app": pachdName,
			},
			Ports: []api.ServicePort{
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
			},
		},
	}
}

// EtcdRc returns an etcd replication controller.
func EtcdRc(hostPath string) *api.ReplicationController {
	replicas := int32(1)
	return &api.ReplicationController{
		TypeMeta: unversioned.TypeMeta{
			Kind:       "ReplicationController",
			APIVersion: "v1",
		},
		ObjectMeta: api.ObjectMeta{
			Name:   etcdName,
			Labels: labels(etcdName),
		},
		Spec: api.ReplicationControllerSpec{
			Replicas: &replicas,
			Selector: map[string]string{
				"app": etcdName,
			},
			Template: &api.PodTemplateSpec{
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
								"--bind-addr=0.0.0.0:2379",
								"--data-dir=/var/data/etcd",
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
						},
					},
					Volumes: []api.Volume{
						{
							Name: "etcd-storage",
							VolumeSource: api.VolumeSource{
								HostPath: &api.HostPathVolumeSource{
									Path: filepath.Join(hostPath, "etcd"),
								},
							},
						},
					},
				},
			},
		},
	}
}

// EtcdService returns an etcd service.
func EtcdService() *api.Service {
	return &api.Service{
		TypeMeta: unversioned.TypeMeta{
			Kind:       "Service",
			APIVersion: "v1",
		},
		ObjectMeta: api.ObjectMeta{
			Name:   etcdName,
			Labels: labels(etcdName),
		},
		Spec: api.ServiceSpec{
			Selector: map[string]string{
				"app": etcdName,
			},
			Ports: []api.ServicePort{
				{
					Port: 2379,
					Name: "client-port",
				},
				{
					Port: 2380,
					Name: "peer-port",
				},
			},
		},
	}
}

// RethinkRc returns a rethinkdb replication controller.
func RethinkRc(backend backend, volume string, hostPath string, rethinkdbCacheSize string) *api.ReplicationController {
	replicas := int32(1)
	rethinkCacheQuantity := resource.MustParse(rethinkdbCacheSize)
	containerFootprint := rethinkCacheQuantity.Copy()
	containerFootprint.Add(resource.MustParse("256M"))
	spec := &api.ReplicationController{
		TypeMeta: unversioned.TypeMeta{
			Kind:       "ReplicationController",
			APIVersion: "v1",
		},
		ObjectMeta: api.ObjectMeta{
			Name:   rethinkName,
			Labels: labels(rethinkName),
		},
		Spec: api.ReplicationControllerSpec{
			Replicas: &replicas,
			Selector: map[string]string{
				"app": rethinkName,
			},
			Template: &api.PodTemplateSpec{
				ObjectMeta: api.ObjectMeta{
					Name:   rethinkName,
					Labels: labels(rethinkName),
				},
				Spec: api.PodSpec{
					Containers: []api.Container{
						{
							Name:  rethinkName,
							Image: rethinkImage,
							//TODO figure out how to get a cluster of these to talk to each other
							Command: []string{"rethinkdb"},
							Args: []string{
								"-d", "/var/rethinkdb/data",
								"--bind", "all",
								"--cache-size", strconv.FormatInt(rethinkCacheQuantity.ScaledValue(resource.Mega), 10),
							},
							Ports: []api.ContainerPort{
								{
									ContainerPort: 8080,
									Name:          "admin-port",
								},
								{
									ContainerPort: 28015,
									Name:          "driver-port",
								},
								{
									ContainerPort: 29015,
									Name:          "cluster-port",
								},
							},
							VolumeMounts: []api.VolumeMount{
								{
									Name:      "rethink-storage",
									MountPath: "/var/rethinkdb/",
								},
							},
							ImagePullPolicy: "IfNotPresent",
							Resources: api.ResourceRequirements{
								Requests: api.ResourceList{
									api.ResourceMemory: *containerFootprint,
								},
							},
						},
					},
					Volumes: []api.Volume{
						{
							Name: "rethink-storage",
							VolumeSource: api.VolumeSource{
								PersistentVolumeClaim: &api.PersistentVolumeClaimVolumeSource{
									ClaimName: rethinkVolumeClaimName,
								},
							},
						},
					},
				},
			},
		},
	}
	return spec
}

// RethinkService returns a rethinkdb service.
func RethinkService() *api.Service {
	return &api.Service{
		TypeMeta: unversioned.TypeMeta{
			Kind:       "Service",
			APIVersion: "v1",
		},
		ObjectMeta: api.ObjectMeta{
			Name:   rethinkName,
			Labels: labels(rethinkName),
		},
		Spec: api.ServiceSpec{
			Type: api.ServiceTypeNodePort,
			Selector: map[string]string{
				"app": rethinkName,
			},
			Ports: []api.ServicePort{
				{
					Port:     8080,
					Name:     "admin-port",
					NodePort: 32080,
				},
				{
					Port:     28015,
					Name:     "driver-port",
					NodePort: 32081,
				},
				{
					Port:     29015,
					Name:     "cluster-port",
					NodePort: 32082,
				},
			},
		},
	}
}

// InitJob returns a pachd-init job.
func InitJob(version string) *extensions.Job {
	image := pachdImage
	if version != "" {
		image += ":" + version
	}
	return &extensions.Job{
		TypeMeta: unversioned.TypeMeta{
			Kind:       "Job",
			APIVersion: "extensions/v1beta1",
		},
		ObjectMeta: api.ObjectMeta{
			Name:   initName,
			Labels: labels(initName),
		},
		Spec: extensions.JobSpec{
			Selector: &extensions.LabelSelector{
				MatchLabels: labels(initName),
			},
			Template: api.PodTemplateSpec{
				ObjectMeta: api.ObjectMeta{
					Name:   initName,
					Labels: labels(initName),
				},
				Spec: api.PodSpec{
					Containers: []api.Container{
						{
							Name:  initName,
							Image: image,
							Env: []api.EnvVar{
								{
									Name:  "PACH_ROOT",
									Value: "/pach",
								},
								{
									Name:  "INIT",
									Value: "true",
								},
							},
							ImagePullPolicy: "IfNotPresent",
						},
					},
					RestartPolicy: "OnFailure",
				},
			},
		},
	}
}

func JobShimInitJob(version string) *extensions.Job {
	image := jobShimImage
	if version != "" {
		image += ":" + version
	}

	var volumes []api.Volume
	var volumeMounts []api.VolumeMount
	volumes = append(volumes, api.Volume{
		Name: "pach-bin",
		VolumeSource: api.VolumeSource{
			HostPath: &api.HostPathVolumeSource{
				Path: "/var/pach/bin",
			},
		},
	})
	volumeMounts = append(volumeMounts, api.VolumeMount{
		Name:      "pach-bin",
		MountPath: "/pach-bin",
	})

	return &extensions.Job{
		TypeMeta: unversioned.TypeMeta{
			Kind:       "Job",
			APIVersion: "extensions/v1beta1",
		},
		ObjectMeta: api.ObjectMeta{
			Name:   jobShimInitName,
			Labels: labels(jobShimInitName),
		},
		Spec: extensions.JobSpec{
			Selector: &extensions.LabelSelector{
				MatchLabels: labels(jobShimInitName),
			},
			Template: api.PodTemplateSpec{
				ObjectMeta: api.ObjectMeta{
					Name:   jobShimInitName,
					Labels: labels(jobShimInitName),
				},
				Spec: api.PodSpec{
					Containers: []api.Container{
						{
							Name:            "init",
							Image:           image,
							Command:         []string{"/pach/job-shim.sh"},
							ImagePullPolicy: api.PullIfNotPresent,
							VolumeMounts:    volumeMounts,
						},
					},
					RestartPolicy: "OnFailure",
					Volumes:       volumes,
				},
			},
		},
	}
}

// AmazonSecret creates an amazon secret with the following parameters:
//   bucket - S3 bucket name
//   id     - AWS access key id
//   secret - AWS secret access key
//   token  - AWS access token
//   region - AWS region
func AmazonSecret(bucket string, id string, secret string, token string, region string) *api.Secret {
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
			"bucket": []byte(bucket),
			"id":     []byte(id),
			"secret": []byte(secret),
			"token":  []byte(token),
			"region": []byte(region),
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

// RethinkVolume creates a persistent volume with a backend
// (local, amazon, google), a name, and a size in gigabytes.
func RethinkVolume(backend backend, hostPath string, name string, size int) *api.PersistentVolume {
	spec := &api.PersistentVolume{
		TypeMeta: unversioned.TypeMeta{
			Kind:       "PersistentVolume",
			APIVersion: "v1",
		},
		ObjectMeta: api.ObjectMeta{
			Name:   rethinkVolumeName,
			Labels: volumeLabels(rethinkVolumeName),
		},
		Spec: api.PersistentVolumeSpec{
			Capacity: map[api.ResourceName]resource.Quantity{
				"storage": resource.MustParse(fmt.Sprintf("%vGi", size)),
			},
			AccessModes:                   []api.PersistentVolumeAccessMode{api.ReadWriteOnce},
			PersistentVolumeReclaimPolicy: api.PersistentVolumeReclaimRetain,
		},
	}

	switch backend {
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
	case localBackend:
		spec.Spec.PersistentVolumeSource = api.PersistentVolumeSource{
			HostPath: &api.HostPathVolumeSource{
				Path: filepath.Join(hostPath, "rethink"),
			},
		}
	default:
		panic("cannot generate volume spec for unknown backend")
	}

	return spec
}

// RethinkVolumeClaim creates a persistent volume claim with a size in gigabytes.
func RethinkVolumeClaim(size int) *api.PersistentVolumeClaim {
	return &api.PersistentVolumeClaim{
		TypeMeta: unversioned.TypeMeta{
			Kind:       "PersistentVolumeClaim",
			APIVersion: "v1",
		},
		ObjectMeta: api.ObjectMeta{
			Name:   rethinkVolumeClaimName,
			Labels: volumeLabels(rethinkVolumeClaimName),
		},
		Spec: api.PersistentVolumeClaimSpec{
			Resources: api.ResourceRequirements{
				Requests: map[api.ResourceName]resource.Quantity{
					"storage": resource.MustParse(fmt.Sprintf("%vGi", size)),
				},
			},
			AccessModes: []api.PersistentVolumeAccessMode{api.ReadWriteOnce},
		},
	}
}

// AssetOpts are options that are applicable to all the asset types.
type AssetOpts struct {
	Shards             uint64
	RethinkdbCacheSize string
	Version            string
	LogLevel           string
}

// WriteAssets writes the assets to w.
func WriteAssets(w io.Writer, opts *AssetOpts, backend backend,
	volumeName string, volumeSize int, hostPath string) {
	encoder := codec.NewEncoder(w, jsonEncoderHandle)

	ServiceAccount().CodecEncodeSelf(encoder)
	fmt.Fprintf(w, "\n")

	RethinkVolume(backend, hostPath, volumeName, volumeSize).CodecEncodeSelf(encoder)
	fmt.Fprintf(w, "\n")
	RethinkVolumeClaim(volumeSize).CodecEncodeSelf(encoder)
	fmt.Fprintf(w, "\n")

	EtcdRc(hostPath).CodecEncodeSelf(encoder)
	fmt.Fprintf(w, "\n")
	EtcdService().CodecEncodeSelf(encoder)
	fmt.Fprintf(w, "\n")

	RethinkService().CodecEncodeSelf(encoder)
	fmt.Fprintf(w, "\n")
	RethinkRc(backend, volumeName, hostPath, opts.RethinkdbCacheSize).CodecEncodeSelf(encoder)
	fmt.Fprintf(w, "\n")

	InitJob(opts.Version).CodecEncodeSelf(encoder)
	fmt.Fprintf(w, "\n")
	JobShimInitJob(opts.Version).CodecEncodeSelf(encoder)
	fmt.Fprintf(w, "\n")

	PachdService().CodecEncodeSelf(encoder)
	fmt.Fprintf(w, "\n")
	PachdRc(opts.Shards, backend, hostPath, opts.LogLevel, opts.Version).CodecEncodeSelf(encoder)
	fmt.Fprintf(w, "\n")
}

// WriteLocalAssets writes assets to a local backend.
func WriteLocalAssets(w io.Writer, opts *AssetOpts, hostPath string) {
	WriteAssets(w, opts, localBackend, "", 0, hostPath)
}

// WriteAmazonAssets writes assets to an amazon backend.
func WriteAmazonAssets(w io.Writer, opts *AssetOpts, bucket string, id string, secret string,
	token string, region string, volumeName string, volumeSize int) {
	WriteAssets(w, opts, amazonBackend, volumeName, volumeSize, "")
	encoder := codec.NewEncoder(w, jsonEncoderHandle)
	AmazonSecret(bucket, id, secret, token, region).CodecEncodeSelf(encoder)
	fmt.Fprintf(w, "\n")
}

// WriteGoogleAssets writes assets to a google backend.
func WriteGoogleAssets(w io.Writer, opts *AssetOpts, bucket string, volumeName string, volumeSize int) {
	WriteAssets(w, opts, googleBackend, volumeName, volumeSize, "")
	encoder := codec.NewEncoder(w, jsonEncoderHandle)
	GoogleSecret(bucket).CodecEncodeSelf(encoder)
	fmt.Fprintf(w, "\n")
}

// WriteMicrosoftAssets writes assets to a microsoft backend
func WriteMicrosoftAssets(w io.Writer, opts *AssetOpts, container string, id string, secret string, volumeURI string, volumeSize int) {
	WriteAssets(w, opts, microsoftBackend, volumeURI, volumeSize, "")
	encoder := codec.NewEncoder(w, jsonEncoderHandle)
	MicrosoftSecret(container, id, secret).CodecEncodeSelf(encoder)
	fmt.Fprintf(w, "\n")
}

func labels(name string) map[string]string {
	return map[string]string{
		"app":   name,
		"suite": suite,
	}
}

func volumeLabels(name string) map[string]string {
	return map[string]string{
		"app":   name,
		"suite": volumeSuite,
	}
}

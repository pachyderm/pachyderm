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
	"k8s.io/kubernetes/pkg/api/unversioned"
	api "k8s.io/kubernetes/pkg/api/v1"
)

var (
	suite               = "pachyderm"
	pachdImage          = "pachyderm/pachd"
	etcdImage           = "quay.io/coreos/etcd:v3.1.2"
	serviceAccountName  = "pachyderm"
	etcdName            = "etcd"
	pachdName           = "pachd"
	minioSecretName     = "minio-secret"
	amazonSecretName    = "amazon-secret"
	googleSecretName    = "google-secret"
	microsoftSecretName = "microsoft-secret"
	trueVal             = true
	jsonEncoderHandle   = &codec.JsonHandle{
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
	// BlockCacheSize is the amount of memory each PachD node allocates towards
	// its cache of PFS blocks.
	BlockCacheSize string
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

// PachdRc returns a pachd replication controller.
func PachdRc(opts *AssetOpts, objectStoreBackend backend, hostPath string) *api.ReplicationController {
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
	switch objectStoreBackend {
	case localBackend:
		volumes[0].HostPath = &api.HostPathVolumeSource{
			Path: filepath.Join(hostPath, "pachd"),
		}
	case minioBackend:
		backendEnvVar = server.MinioBackendEnvVar
		volumes[0].HostPath = &api.HostPathVolumeSource{
			Path: filepath.Join(hostPath, "pachd"),
		}
		volumes = append(volumes, api.Volume{
			Name: minioSecretName,
			VolumeSource: api.VolumeSource{
				Secret: &api.SecretVolumeSource{
					SecretName: minioSecretName,
				},
			},
		})
		volumeMounts = append(volumeMounts, api.VolumeMount{
			Name:      minioSecretName,
			MountPath: "/" + minioSecretName,
		})
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
									Value: fmt.Sprintf("%d", opts.PachdShards),
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
									Name:  "WORKER_IMAGE",
									Value: fmt.Sprintf("pachyderm/worker:%s", opts.Version),
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
							},
							VolumeMounts: volumeMounts,
							SecurityContext: &api.SecurityContext{
								Privileged: &trueVal, // god is this dumb
							},
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
								"--listen-client-urls=http://0.0.0.0:2379",
								"--advertise-client-urls=http://0.0.0.0:2379",
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
func EtcdService(local bool) *api.Service {
	var clientNodePort int32
	serviceType := api.ServiceTypeNodePort
	if local {
		clientNodePort = 32379
		serviceType = api.ServiceTypeNodePort
	}
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
			Type: serviceType,
			Selector: map[string]string{
				"app": etcdName,
			},
			Ports: []api.ServicePort{
				{
					Port:     2379,
					Name:     "client-port",
					NodePort: clientNodePort,
				},
				{
					Port: 2380,
					Name: "peer-port",
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

// WriteAssets writes the assets to w.
func WriteAssets(w io.Writer, opts *AssetOpts, objectStoreBackend backend,
	persistentDiskBackend backend, volumeNames []string, volumeSize int,
	hostPath string) error {
	// If either backend is "local", both must be "local"
	if (persistentDiskBackend == localBackend || objectStoreBackend == localBackend) &&
		persistentDiskBackend != objectStoreBackend {
		return fmt.Errorf("if either persistentDiskBackend or objectStoreBackend "+
			"is \"local\", both must be \"local\", but persistentDiskBackend==%d, \n"+
			"and objectStoreBackend==%d", persistentDiskBackend, objectStoreBackend)
	}
	encoder := codec.NewEncoder(w, jsonEncoderHandle)

	ServiceAccount().CodecEncodeSelf(encoder)
	fmt.Fprintf(w, "\n")

	EtcdRc(hostPath).CodecEncodeSelf(encoder)
	fmt.Fprintf(w, "\n")
	EtcdService(objectStoreBackend == localBackend).CodecEncodeSelf(encoder)
	fmt.Fprintf(w, "\n")

	PachdService().CodecEncodeSelf(encoder)
	fmt.Fprintf(w, "\n")
	PachdRc(opts, objectStoreBackend, hostPath).CodecEncodeSelf(encoder)
	fmt.Fprintf(w, "\n")
	return nil
}

// WriteLocalAssets writes assets to a local backend.
func WriteLocalAssets(w io.Writer, opts *AssetOpts, hostPath string) error {
	return WriteAssets(w, opts, localBackend, localBackend, nil, 1 /* = volume size (gb) */, hostPath)
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
			if err := WriteAssets(w, opts, minioBackend, amazonBackend, strings.Split(args[0], ","), volumeSize, ""); err != nil {
				return err
			}
		case "google":
			if err := WriteAssets(w, opts, minioBackend, googleBackend, strings.Split(args[0], ","), volumeSize, ""); err != nil {
				return err
			}
		case "azure":
			if err := WriteAssets(w, opts, minioBackend, microsoftBackend, strings.Split(args[0], ","), volumeSize, ""); err != nil {
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
	token string, region string, volumeNames []string, volumeSize int) error {
	if err := WriteAssets(w, opts, amazonBackend, amazonBackend, volumeNames, volumeSize, ""); err != nil {
		return err
	}
	encoder := codec.NewEncoder(w, jsonEncoderHandle)
	AmazonSecret(bucket, id, secret, token, region).CodecEncodeSelf(encoder)
	fmt.Fprintf(w, "\n")
	return nil
}

// WriteGoogleAssets writes assets to a google backend.
func WriteGoogleAssets(w io.Writer, opts *AssetOpts, bucket string, volumeNames []string, volumeSize int) error {
	if err := WriteAssets(w, opts, googleBackend, googleBackend, volumeNames, volumeSize, ""); err != nil {
		return err
	}
	encoder := codec.NewEncoder(w, jsonEncoderHandle)
	GoogleSecret(bucket).CodecEncodeSelf(encoder)
	fmt.Fprintf(w, "\n")
	return nil
}

// WriteMicrosoftAssets writes assets to a microsoft backend
func WriteMicrosoftAssets(w io.Writer, opts *AssetOpts, container string, id string, secret string, volumeURIs []string, volumeSize int) error {
	if err := WriteAssets(w, opts, microsoftBackend, microsoftBackend, volumeURIs, volumeSize, ""); err != nil {
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

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
)

var (
	suite                   = "pachyderm"
	pachdImage              = "pachyderm/pachd"
	etcdImage               = "quay.io/coreos/etcd:v3.1.2"
	serviceAccountName      = "pachyderm"
	etcdHeadlessServiceName = "etcd-headless"
	etcdVolumeName          = "etcd-claim"
	etcdVolumeClaimName     = "etcd-storage"
	etcdName                = "etcd"
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
func PachdRc(shards uint64, backend backend, hostPath string, logLevel string, version string, metrics bool) *api.ReplicationController {
	image := pachdImage
	if version != "" {
		image += ":" + version
	}
	// we turn metrics off if we dont have a static version
	// this prevents dev clusters from reporting metrics
	if version == deploy.DevVersionTag {
		metrics = false
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
	switch backend {
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
									Name:  "WORKER_IMAGE",
									Value: fmt.Sprintf("pachyderm/worker:%s", version),
								},
								{
									Name:  "WORKER_IMAGE_PULL_POLICY",
									Value: "IfNotPresent",
								},
								{
									Name:  "PACHD_VERSION",
									Value: version,
								},
								{
									Name:  "METRICS",
									Value: strconv.FormatBool(metrics),
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
								// HostPath: &api.HostPathVolumeSource{
								// 	Path: filepath.Join(hostPath, "etcd"),
								// },
								PersistentVolumeClaim: &api.PersistentVolumeClaimVolumeSource{
									ClaimName: etcdVolumeClaimName,
								},
							},
						},
					},
				},
			},
		},
	}
}

// WriteEtcdVolumes creates 'shards' persistent volumes, either backed by IAAS persistent volumes (EBS volumes for amazon, GCP volumes for Google, etc)
// or local volumes (if 'backend' == 'local'). All volumes are created with size 'size'.
func WriteEtcdVolumes(w io.Writer, persistentDiskBackend backend,
	opts *AssetOpts, hostPath string, names []string,
	size int) error {
	if persistentDiskBackend != localBackend && len(names) < opts.EtcdNodes {
		return fmt.Errorf("could not create non-local rethink cluster with %d shards, as there are only %d external volumes", opts.EtcdNodes, len(names))
	}
	encoder := codec.NewEncoder(w, jsonEncoderHandle)
	for i := 0; i < opts.EtcdNodes; i++ {
		spec := &api.PersistentVolume{
			TypeMeta: unversioned.TypeMeta{
				Kind:       "PersistentVolume",
				APIVersion: "v1",
			},
			ObjectMeta: api.ObjectMeta{
				Name:   fmt.Sprintf("%s-%d", etcdVolumeName, i),
				Labels: labels(etcdVolumeName),
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
					VolumeID: names[i],
				},
			}
		case googleBackend:
			spec.Spec.PersistentVolumeSource = api.PersistentVolumeSource{
				GCEPersistentDisk: &api.GCEPersistentDiskVolumeSource{
					FSType: "ext4",
					PDName: names[i],
				},
			}
		case microsoftBackend:
			dataDiskURI := names[i]
			split := strings.Split(names[i], "/")
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
					Path: filepath.Join(hostPath, fmt.Sprintf("etcd-%d", i)),
				},
			}
		default:
			return fmt.Errorf("cannot generate volume spec for unknown backend \"%v\"", persistentDiskBackend)
		}
		spec.CodecEncodeSelf(encoder)
		fmt.Fprintf(w, "\n")
	}
	return nil
}

// EtcdNodePortService returns a NodePort etcd service. This will let non-etcd
// pods talk to etcd
func EtcdNodePortService(local bool) *api.Service {
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
			},
		},
	}
}

// EtcdHeadlessService returns a headless etcd service, which is only for DNS
// resolution.
func EtcdHeadlessService() *api.Service {
	return &api.Service{
		TypeMeta: unversioned.TypeMeta{
			Kind:       "Service",
			APIVersion: "v1",
		},
		ObjectMeta: api.ObjectMeta{
			Name:   etcdHeadlessServiceName,
			Labels: labels(etcdName),
		},
		Spec: api.ServiceSpec{
			Selector: map[string]string{
				"app": etcdName,
			},
			ClusterIP: "None",
			Ports: []api.ServicePort{
				{
					Name: "peer-port",
					Port: 2380,
				},
			},
		},
	}
}

// EtcdStatefulSet returns a stateful set that manages an etcd cluster
func EtcdStatefulSet(opts *AssetOpts, diskSpace int) interface{} {
	initialCluster := make([]string, 0, opts.EtcdNodes)
	for i := 0; i < opts.EtcdNodes; i++ {
		url := fmt.Sprintf("http://etcd-%d.etcd-headless.default.svc.cluster.local:2380", i)
		// TODO(msteffen): give each node a name using the downward API, and use it
		// here instead of "default"
		initialCluster = append(initialCluster, fmt.Sprintf("etcd-%d=%s", i, url))
	}
	fmt.Println("opts.EtcdNodes: ", opts.EtcdNodes)
	fmt.Println("initialCluster: ", initialCluster)
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
		"--initial-advertise-peer-urls=http://${ETCD_NAME}.etcd-headless.default.svc.cluster.local:2380",
		"--initial-cluster=" + strings.Join(initialCluster, ","),
	}
	for i, str := range etcdCmd {
		etcdCmd[i] = fmt.Sprintf("\"%s\"", str)
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
								// TODO figure out resource requirements
								},
							},
						},
					},
				},
			},
			"volumeClaimTemplates": []interface{}{
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

// AssetOpts are options that are applicable to all the asset types.
type AssetOpts struct {
	PachdShards uint64
	Version     string
	LogLevel    string
	Metrics     bool
	EtcdNodes   int
}

// WriteAssets writes the assets to w.
func WriteAssets(w io.Writer, opts *AssetOpts, backend backend,
	volumeNames []string, volumeSize int, hostPath string) error {
	encoder := codec.NewEncoder(w, jsonEncoderHandle)

	ServiceAccount().CodecEncodeSelf(encoder)
	fmt.Fprintf(w, "\n")

	WriteEtcdVolumes(w, backend, opts, hostPath, volumeNames, volumeSize)
	if opts.EtcdNodes > 1 {
		EtcdHeadlessService().CodecEncodeSelf(encoder)
		fmt.Fprintf(w, "\n")
		encoder.Encode(EtcdStatefulSet(opts, volumeSize))
		fmt.Fprintf(w, "\n")
	} else {
		EtcdRc(hostPath).CodecEncodeSelf(encoder)
		fmt.Fprintf(w, "\n")
	}
	EtcdNodePortService(backend == localBackend).CodecEncodeSelf(encoder)
	fmt.Fprintf(w, "\n")

	PachdService().CodecEncodeSelf(encoder)
	fmt.Fprintf(w, "\n")
	PachdRc(opts.PachdShards, backend, hostPath, opts.LogLevel, opts.Version, opts.Metrics).CodecEncodeSelf(encoder)
	fmt.Fprintf(w, "\n")
	return nil
}

// WriteLocalAssets writes assets to a local backend.
func WriteLocalAssets(w io.Writer, opts *AssetOpts, hostPath string) error {
	return WriteAssets(w, opts, localBackend, nil, 1 /* = volume size (gb) */, hostPath)
}

// WriteMinioAssets writes assets to an s3 backend.
func WriteMinioAssets(w io.Writer, opts *AssetOpts, hostPath string, bucket string, id string,
	secret string, endpoint string, secure bool) error {
	if err := WriteAssets(w, opts, minioBackend, nil, 1 /* = volume size (gb) */, hostPath); err != nil {
		return err
	}
	encoder := codec.NewEncoder(w, jsonEncoderHandle)
	MinioSecret(bucket, id, secret, endpoint, secure).CodecEncodeSelf(encoder)
	fmt.Fprintf(w, "\n")
	return nil
}

// WriteAmazonAssets writes assets to an amazon backend.
func WriteAmazonAssets(w io.Writer, opts *AssetOpts, bucket string, id string, secret string,
	token string, region string, volumeNames []string, volumeSize int) error {
	if err := WriteAssets(w, opts, amazonBackend, volumeNames, volumeSize, ""); err != nil {
		return err
	}
	encoder := codec.NewEncoder(w, jsonEncoderHandle)
	AmazonSecret(bucket, id, secret, token, region).CodecEncodeSelf(encoder)
	fmt.Fprintf(w, "\n")
	return nil
}

// WriteGoogleAssets writes assets to a google backend.
func WriteGoogleAssets(w io.Writer, opts *AssetOpts, bucket string, volumeNames []string, volumeSize int) error {
	if err := WriteAssets(w, opts, googleBackend, volumeNames, volumeSize, ""); err != nil {
		return err
	}
	encoder := codec.NewEncoder(w, jsonEncoderHandle)
	GoogleSecret(bucket).CodecEncodeSelf(encoder)
	fmt.Fprintf(w, "\n")
	return nil
}

// WriteMicrosoftAssets writes assets to a microsoft backend
func WriteMicrosoftAssets(w io.Writer, opts *AssetOpts, container string, id string, secret string, volumeURIs []string, volumeSize int) error {
	if err := WriteAssets(w, opts, microsoftBackend, volumeURIs, volumeSize, ""); err != nil {
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

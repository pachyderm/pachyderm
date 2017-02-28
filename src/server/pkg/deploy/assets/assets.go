package assets

import (
	"fmt"
	"io"
	"net"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/pachyderm/pachyderm/src/client"
	"github.com/pachyderm/pachyderm/src/server/pfs/server"
	"github.com/pachyderm/pachyderm/src/server/pkg/deploy"
	"github.com/ugorji/go/codec"
	"k8s.io/kubernetes/pkg/api/resource"
	"k8s.io/kubernetes/pkg/api/unversioned"
	api "k8s.io/kubernetes/pkg/api/v1"
	extensions "k8s.io/kubernetes/pkg/apis/extensions/v1beta1"
)

var (
	suite                       = "pachyderm"
	pachdImage                  = "pachyderm/pachd"
	etcdImage                   = "gcr.io/google_containers/etcd:2.0.12"
	rethinkImage                = "rethinkdb:2.3.2"
	rethinkNonCacheMemFootprint = resource.MustParse("256M") // Amount of memory needed by rethink beyond the cache
	serviceAccountName          = "pachyderm"
	etcdName                    = "etcd"
	pachdName                   = "pachd"
	rethinkControllerName       = "rethink" // Used by both the RethinkDB Stateful Set and ReplicationController (whichever is enabled)
	rethinkServiceName          = "rethink"
	rethinkHeadlessName         = "rethink-headless" // headless service; give Rethink pods consistent DNS addresses
	rethinkVolumeName           = "rethink-volume"
	rethinkVolumeClaimName      = "rethink-volume-claim"
	minioSecretName             = "minio-secret"
	amazonSecretName            = "amazon-secret"
	googleSecretName            = "google-secret"
	microsoftSecretName         = "microsoft-secret"
	initName                    = "pachd-init"
	trueVal                     = true
	jsonEncoderHandle           = &codec.JsonHandle{
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
									Value: strconv.FormatBool(metrics),
								},
								{
									Name:  "LOG_LEVEL",
									Value: logLevel,
								},
								{
									Name:  client.PPSLeasePeriodSecsEnv,
									Value: "30",
								},
								{
									Name:  client.PPSHeartbeatSecsEnv,
									Value: "10",
								},
								{
									Name:  client.PPSMaxHeartbeatRetriesEnv,
									Value: "3",
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
func RethinkRc(volume string, rethinkdbCacheSize string) *api.ReplicationController {
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
			Name:   rethinkControllerName,
			Labels: labels(rethinkControllerName),
		},
		Spec: api.ReplicationControllerSpec{
			Replicas: &replicas,
			Selector: map[string]string{
				"app": rethinkControllerName,
			},
			Template: &api.PodTemplateSpec{
				ObjectMeta: api.ObjectMeta{
					Name:   rethinkControllerName,
					Labels: labels(rethinkControllerName),
				},
				Spec: api.PodSpec{
					Containers: []api.Container{
						{
							Name:  rethinkControllerName,
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

// RethinkStatefulSet returns a rethinkdb stateful set
func RethinkStatefulSet(shards int, diskSpace int, cacheSize string) interface{} {
	rethinkCacheQuantity := resource.MustParse(cacheSize)
	containerFootprint := rethinkCacheQuantity.Copy()
	containerFootprint.Add(rethinkNonCacheMemFootprint)
	// As of Oct 24 2016, the Kubernetes client does not include structs for Stateful Set, so we generate the kubernetes
	// manifest using raw json.

	// Stateful Set config:
	return map[string]interface{}{
		"apiVersion": "apps/v1beta1",
		"kind":       "StatefulSet",
		"metadata": map[string]interface{}{
			"name":              rethinkControllerName,
			"creationTimestamp": nil,
			"labels":            labels(rethinkServiceName),
		},
		"spec": map[string]interface{}{
			// Effectively configures a RC
			"serviceName": rethinkHeadlessName,
			"replicas":    shards,
			"selector": map[string]interface{}{
				"matchLabels": labels(rethinkControllerName),
			},

			// pod template
			"template": map[string]interface{}{
				"metadata": map[string]interface{}{
					"name":              rethinkControllerName,
					"creationTimestamp": nil,
					"annotations":       map[string]string{"pod.alpha.kubernetes.io/initialized": "true"},
					"labels":            labels(rethinkControllerName),
				},
				"spec": map[string]interface{}{
					"containers": []interface{}{
						map[string]interface{}{
							"name":    rethinkControllerName,
							"image":   rethinkImage,
							"command": []string{"rethinkdb"},
							"args": []string{
								"-d", "/var/rethinkdb/data",
								"--bind", "all",
								"--cache-size", strconv.FormatInt(rethinkCacheQuantity.ScaledValue(resource.Mega), 10),
								"--join", net.JoinHostPort("rethink-0.rethink-headless.default.svc.cluster.local", "29015"),
							},
							"ports": []interface{}{
								map[string]interface{}{
									"containerPort": 8080,
									"name":          "admin-port",
								},
								map[string]interface{}{
									"containerPort": 28015,
									"name":          "driver-port",
								},
								map[string]interface{}{
									"containerPort": 29015,
									"name":          "cluster-port",
								},
							},
							"volumeMounts": []interface{}{
								map[string]interface{}{
									"name":      rethinkVolumeClaimName,
									"mountPath": "/var/rethinkdb/",
								},
							},
							"imagePullPolicy": "IfNotPresent",
							"resources": map[string]interface{}{
								"requests": map[string]interface{}{
									"memory": containerFootprint.String(),
								},
							},
						},
					},
				},
			},
			"volumeClaimTemplates": []interface{}{
				map[string]interface{}{
					"metadata": map[string]interface{}{
						"name":   rethinkVolumeClaimName,
						"labels": labels(rethinkVolumeName),
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

// RethinkHeadlessService returns a headless rethinkdb service, which is only for DNS resolution.
func RethinkHeadlessService() *api.Service {
	return &api.Service{
		TypeMeta: unversioned.TypeMeta{
			Kind:       "Service",
			APIVersion: "v1",
		},
		ObjectMeta: api.ObjectMeta{
			Name:   rethinkHeadlessName,
			Labels: labels(rethinkServiceName),
		},
		Spec: api.ServiceSpec{
			Selector:  labels(rethinkControllerName),
			ClusterIP: "None",
			Ports: []api.ServicePort{
				{
					Port: 29015,
					Name: "cluster-port",
				},
			},
		},
	}
}

// RethinkNodeportService returns a rethinkdb NodePort service.
func RethinkNodeportService(opts *AssetOpts) *api.Service {
	serviceDef := &api.Service{
		TypeMeta: unversioned.TypeMeta{
			Kind:       "Service",
			APIVersion: "v1",
		},
		ObjectMeta: api.ObjectMeta{
			Name:   rethinkServiceName,
			Labels: labels(rethinkServiceName),
		},
		Spec: api.ServiceSpec{
			Type:     api.ServiceTypeNodePort,
			Selector: labels(rethinkControllerName),
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
			},
		},
	}

	if !opts.DeployRethinkAsStatefulSet {
		serviceDef.Spec.Ports = append(serviceDef.Spec.Ports, api.ServicePort{
			Port: 29015,
			Name: "cluster-port",
		})
	}
	return serviceDef
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

// WriteRethinkVolumes creates 'shards' persistent volumes, either backed by IAAS persistent volumes (EBS volumes for amazon, GCP volumes for Google, etc)
// or local volumes (if 'backend' == 'local'). All volumes are created with size 'size'.
func WriteRethinkVolumes(w io.Writer, backend backend, shards int, hostPath string, names []string, size int) error {
	if backend != localBackend && backend != minioBackend && len(names) < shards {
		return fmt.Errorf("could not create non-local rethink cluster with %d shards, as there are only %d external volumes", shards, len(names))
	}
	encoder := codec.NewEncoder(w, jsonEncoderHandle)
	for i := 0; i < shards; i++ {
		spec := &api.PersistentVolume{
			TypeMeta: unversioned.TypeMeta{
				Kind:       "PersistentVolume",
				APIVersion: "v1",
			},
			ObjectMeta: api.ObjectMeta{
				Name:   fmt.Sprintf("%s-%d", rethinkVolumeName, i),
				Labels: labels(rethinkVolumeName),
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
					Path: filepath.Join(hostPath, fmt.Sprintf("rethink-%d", i)),
				},
			}
		default:
			return fmt.Errorf("cannot generate volume spec for unknown backend \"%v\"", backend)
		}
		spec.CodecEncodeSelf(encoder)
		fmt.Fprintf(w, "\n")
	}
	return nil
}

// RethinkVolumeClaim creates a persistent volume claim with a size in gigabytes.
//
// Note that if you're controlling RethinkDB as a Stateful Set, this is unneccessary.
// We're only keeping it for backwards compatibility with GKE. Therefore at most one
// persistent volume claim will be created by this function, so it's okay to name it
// statically
func RethinkVolumeClaim(size int) *api.PersistentVolumeClaim {
	return &api.PersistentVolumeClaim{
		TypeMeta: unversioned.TypeMeta{
			Kind:       "PersistentVolumeClaim",
			APIVersion: "v1",
		},
		ObjectMeta: api.ObjectMeta{
			Name:   rethinkVolumeClaimName,
			Labels: labels(rethinkVolumeClaimName),
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
	PachdShards        uint64
	RethinkShards      uint64
	RethinkdbCacheSize string
	Version            string
	LogLevel           string
	Metrics            bool

	// Deploy single-node rethink managed by a RC, rather than a multi-node,
	// highly-available Stateful Set. This will be necessary until GKE supports Stateful Set
	DeployRethinkAsStatefulSet bool
}

// WriteAssets writes the assets to w.
func WriteAssets(w io.Writer, opts *AssetOpts, backend backend,
	volumeNames []string, volumeSize int, hostPath string) error {
	encoder := codec.NewEncoder(w, jsonEncoderHandle)

	ServiceAccount().CodecEncodeSelf(encoder)
	fmt.Fprintf(w, "\n")

	err := WriteRethinkVolumes(w, backend, int(opts.RethinkShards), hostPath, volumeNames, volumeSize)
	if err != nil {
		return err
	}

	EtcdRc(hostPath).CodecEncodeSelf(encoder)
	fmt.Fprintf(w, "\n")
	EtcdService().CodecEncodeSelf(encoder)
	fmt.Fprintf(w, "\n")

	RethinkNodeportService(opts).CodecEncodeSelf(encoder)
	fmt.Fprintf(w, "\n")
	if opts.DeployRethinkAsStatefulSet {
		encoder.Encode(RethinkStatefulSet(int(opts.RethinkShards), volumeSize, opts.RethinkdbCacheSize))
		fmt.Fprintf(w, "\n")
		RethinkHeadlessService().CodecEncodeSelf(encoder)
	} else {
		if backend != localBackend && backend != minioBackend && len(volumeNames) != 1 {
			return fmt.Errorf("RethinkDB can only be managed by a ReplicationController as a single instance, but recieved %d volumes", len(volumeNames))
		}
		RethinkVolumeClaim(volumeSize).CodecEncodeSelf(encoder)
		volumeName := ""
		if backend != localBackend && backend != minioBackend {
			volumeName = volumeNames[0]
		}
		RethinkRc(volumeName, opts.RethinkdbCacheSize).CodecEncodeSelf(encoder)
	}
	fmt.Fprintf(w, "\n")

	InitJob(opts.Version).CodecEncodeSelf(encoder)
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

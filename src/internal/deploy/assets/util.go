package assets

import (
	"fmt"
	"path"
	"strings"

	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	v1 "k8s.io/api/core/v1"
	storagev1 "k8s.io/api/storage/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

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

func makeStorageClass(opts *AssetOpts, backend Backend, storageClassName string, storageClassLabels map[string]string) *storagev1.StorageClass {
	allowVolumeExpansion := true
	sc := &storagev1.StorageClass{
		TypeMeta: metav1.TypeMeta{
			Kind:       "StorageClass",
			APIVersion: "storage.k8s.io/v1",
		},
		ObjectMeta:           objectMeta(storageClassName, storageClassLabels, nil, opts.Namespace),
		AllowVolumeExpansion: &allowVolumeExpansion,
	}
	switch backend {
	case GoogleBackend:
		sc.Provisioner = "kubernetes.io/gce-pd"
		sc.Parameters = map[string]string{"type": "pd-ssd"}
	case AmazonBackend:
		sc.Provisioner = "kubernetes.io/aws-ebs"
		sc.Parameters = map[string]string{"type": "gp2"}
	case LocalBackend:
		sc.Provisioner = "kubernetes.io/no-provisioner"
		bindingMode := storagev1.VolumeBindingWaitForFirstConsumer
		sc.VolumeBindingMode = &bindingMode
	default:
		return nil
	}
	return sc
}

func makePersistentVolume(opts *AssetOpts, persistentDiskBackend Backend, hostPath string, name string, size int, volumeName string, volumeLabels map[string]string) (*v1.PersistentVolume, error) {
	spec := &v1.PersistentVolume{
		TypeMeta: metav1.TypeMeta{
			Kind:       "PersistentVolume",
			APIVersion: "v1",
		},
		ObjectMeta: objectMeta(volumeName, volumeLabels, nil, opts.Namespace),
		Spec: v1.PersistentVolumeSpec{
			Capacity: map[v1.ResourceName]resource.Quantity{
				"storage": resource.MustParse(fmt.Sprintf("%vGi", size)),
			},
			AccessModes:                   []v1.PersistentVolumeAccessMode{v1.ReadWriteOnce},
			PersistentVolumeReclaimPolicy: v1.PersistentVolumeReclaimRetain,
		},
	}

	switch persistentDiskBackend {
	case AmazonBackend:
		spec.Spec.PersistentVolumeSource = v1.PersistentVolumeSource{
			AWSElasticBlockStore: &v1.AWSElasticBlockStoreVolumeSource{
				FSType:   "ext4",
				VolumeID: name,
			},
		}
	case GoogleBackend:
		spec.Spec.PersistentVolumeSource = v1.PersistentVolumeSource{
			GCEPersistentDisk: &v1.GCEPersistentDiskVolumeSource{
				FSType: "ext4",
				PDName: name,
			},
		}
	case MicrosoftBackend:
		dataDiskURI := name
		split := strings.Split(name, "/")
		diskName := split[len(split)-1]

		spec.Spec.PersistentVolumeSource = v1.PersistentVolumeSource{
			AzureDisk: &v1.AzureDiskVolumeSource{
				DiskName:    diskName,
				DataDiskURI: dataDiskURI,
			},
		}
	case MinioBackend:
		fallthrough
	case LocalBackend:
		pathType := v1.HostPathDirectoryOrCreate
		spec.Spec.PersistentVolumeSource = v1.PersistentVolumeSource{
			HostPath: &v1.HostPathVolumeSource{
				Path: path.Join(hostPath, volumeName),
				Type: &pathType,
			},
		}
	default:
		return nil, errors.Errorf("cannot generate volume spec for unknown backend \"%v\"", persistentDiskBackend)
	}
	return spec, nil
}

func makeVolumeClaim(opts *AssetOpts, size int, volumeName string, volumeClaimName string, volumeClaimLabels map[string]string) *v1.PersistentVolumeClaim {
	return &v1.PersistentVolumeClaim{
		TypeMeta: metav1.TypeMeta{
			Kind:       "PersistentVolumeClaim",
			APIVersion: "v1",
		},
		ObjectMeta: objectMeta(volumeClaimName, volumeClaimLabels, nil, opts.Namespace),
		Spec: v1.PersistentVolumeClaimSpec{
			Resources: v1.ResourceRequirements{
				Requests: map[v1.ResourceName]resource.Quantity{
					"storage": resource.MustParse(fmt.Sprintf("%vGi", size)),
				},
			},
			AccessModes: []v1.PersistentVolumeAccessMode{v1.ReadWriteOnce},
			VolumeName:  volumeName,
		},
	}
}

func makeHeadlessService(opts *AssetOpts, name, serviceName string, ports []v1.ServicePort) *v1.Service {
	return &v1.Service{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Service",
			APIVersion: "v1",
		},
		ObjectMeta: objectMeta(serviceName, labels(name), nil, opts.Namespace),
		Spec: v1.ServiceSpec{
			Selector: map[string]string{
				"app": name,
			},
			ClusterIP: "None",
			Ports:     ports,
		},
	}
}

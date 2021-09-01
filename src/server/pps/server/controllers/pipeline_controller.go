package controllers

import (
	"context"
	"fmt"

	ppsv1 "github.com/pachyderm/pachyderm/v2/src/server/pps/server/api/v1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

type PipelineReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=pps.pachyderm.io,resources=pipelines,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=pps.pachyderm.io,resources=pipelines/status,verbs=get;update;patch
func (r *PipelineReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {

	log := log.FromContext(ctx).WithValues("pipeline", req.NamespacedName)
	log.V(1).Info("reconciling pipeline CRD")

	var pipeline ppsv1.Pipeline
	if err := r.Get(ctx, req.NamespacedName, &pipeline); err != nil {
		//log.Info(err, "unable to fetch Pipeline")
		// we'll ignore not-found errors, since they can't be fixed by an immediate
		// requeue (we'll need to wait for a new notification), and we can get them
		// on deleted requests.
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	err := r.Get(ctx, types.NamespacedName{Name: fmt.Sprintf("pipeline-%s-v1", pipeline.Name), Namespace: req.NamespacedName.Namespace}, &appsv1.Deployment{})
	if err != nil && errors.IsNotFound(err) {

		/*err, specCommitID := newPachPipeline(&pipeline, r.PachClient)
		if err != nil {
			return ctrl.Result{}, err
		}*/
		specCommitID := "blah"
		pipelineDeployment := newDeployment(&pipeline, specCommitID)

		if err := controllerutil.SetControllerReference(&pipeline, pipelineDeployment, r.Scheme); err != nil {
			return ctrl.Result{}, err
		}

		log.Info("Creating Deployment: ", pipelineDeployment.Namespace, pipelineDeployment.Name)

		err = r.Create(ctx, pipelineDeployment)
		if err != nil {
			return ctrl.Result{}, err
		}

		pipeline.Status.State = "Running"

		if err := r.Status().Update(ctx, &pipeline); err != nil {
			log.Error(err, "unable to update Pipeline status")
			return ctrl.Result{}, err
		}

	} else if err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil

}

var (
	jobOwnerKey = ".metadata.controller"
	apiGVStr    = ppsv1.GroupVersion.String()
)

func (r *PipelineReconciler) SetupWithManager(mgr ctrl.Manager) error {

	if err := mgr.GetFieldIndexer().IndexField(context.Background(), &appsv1.Deployment{}, jobOwnerKey, func(rawObj client.Object) []string {
		// grab the job object, extract the owner...
		deployment := rawObj.(*appsv1.Deployment)
		owner := metav1.GetControllerOf(deployment)
		if owner == nil {
			return nil
		}
		//Make sure it's a pipeline
		if owner.APIVersion != apiGVStr || owner.Kind != "Pipeline" {
			return nil
		}

		// ...and if so, return it
		return []string{owner.Name}
	}); err != nil {
		return err
	}

	return ctrl.NewControllerManagedBy(mgr).
		For(&ppsv1.Pipeline{}).
		Owns(&appsv1.Deployment{}).
		Complete(r)
}

// newDeployment creates a new Deployment for a Foo resource. It also sets
// the appropriate OwnerReferences on the resource so handleObject can discover
// the Foo resource that 'owns' it.
func newDeployment(pipeline *ppsv1.Pipeline, specCommitID string) *appsv1.Deployment {
	workerImage := "pachyderm/worker:local"
	pachImage := "pachyderm/pachd:local"
	//volumeMounts := ""
	userImage := pipeline.Spec.Transform.Image

	//Need to know storage backend (ie GCS Bucket ) +  Secret

	//What about standby pipelines? - Need a way to get signal from master to spin up again

	//Will controller look at pach etcd for state?

	//Will controller connect to PFS (ie to save pipeline spec)

	//Worker Env
	//TODO Add transform.Env

	commonEnv := []v1.EnvVar{{
		Name:  "PACH_ROOT",
		Value: "/pach",
	}, {
		Name:  "PACH_NAMESPACE",
		Value: "default",
	}, {
		Name:  "STORAGE_BACKEND",
		Value: "LOCAL",
	}, {
		Name:  "POSTGRES_USER",
		Value: "pachyderm",
	}, {
		Name: "POSTGRES_PASSWORD",
		ValueFrom: &v1.EnvVarSource{
			SecretKeyRef: &v1.SecretKeySelector{
				LocalObjectReference: v1.LocalObjectReference{
					Name: "postgres",
				},
				Key: "postgresql-password",
			},
		},
	}, {
		Name:  "POSTGRES_DATABASE",
		Value: "pachyderm",
	}, {
		Name:  "PG_BOUNCER_HOST",
		Value: "pg-bouncer",
	}, {
		Name:  "PG_BOUNCER_PORT",
		Value: "5432",
	}, {
		Name:  "PEER_PORT",
		Value: "1653",
	}, {
		Name:  "PPS_SPEC_COMMIT",
		Value: specCommitID,
	}, {
		Name:  "PPS_PIPELINE_NAME",
		Value: pipeline.Name,
	},
		// These are set explicitly below to prevent kubernetes from setting them to the service host and port.
		{
			Name:  "POSTGRES_PORT",
			Value: "",
		}, {
			Name:  "POSTGRES_HOST",
			Value: "",
		},
	}

	// Set up sidecar env vars
	sidecarEnv := []v1.EnvVar{{
		Name:  "PORT",
		Value: "1650",
	}, {
		Name: "PACHD_POD_NAME",
		ValueFrom: &v1.EnvVarSource{
			FieldRef: &v1.ObjectFieldSelector{
				APIVersion: "v1",
				FieldPath:  "metadata.name",
			},
		},
	}, {
		Name:  "GC_PERCENT",
		Value: "50",
	}, {
		Name:  "STORAGE_UPLOAD_CONCURRENCY_LIMIT",
		Value: "100",
	},
	}
	sidecarEnv = append(sidecarEnv, commonEnv...)

	// Set up worker env vars
	workerEnv := []v1.EnvVar{
		// Set core pach env vars
		{
			Name:  "PACH_IN_WORKER",
			Value: "true",
		},
		// We use Kubernetes' "Downward API" so the workers know their IP
		// addresses, which they will then post on etcd so the job managers
		// can discover the workers.
		{
			Name: "PPS_WORKER_IP",
			ValueFrom: &v1.EnvVarSource{
				FieldRef: &v1.ObjectFieldSelector{
					APIVersion: "v1",
					FieldPath:  "status.podIP",
				},
			},
		},
		// Set the PPS env vars
		{
			Name:  "PPS_ETCD_PREFIX",
			Value: "pachyderm/1.7.0/pachyderm_pps",
		},
		{
			Name: "PPS_POD_NAME",
			ValueFrom: &v1.EnvVarSource{
				FieldRef: &v1.ObjectFieldSelector{
					APIVersion: "v1",
					FieldPath:  "metadata.name",
				},
			},
		},
		{
			Name:  "PPS_WORKER_GRPC_PORT",
			Value: "1080",
		},
	}
	workerEnv = append(workerEnv, commonEnv...)

	envFrom := []v1.EnvFromSource{
		{
			SecretRef: &v1.SecretEnvSource{
				LocalObjectReference: v1.LocalObjectReference{
					Name: "pachyderm-storage-secret",
				},
			},
		},
	}

	initVolumeMounts := []corev1.VolumeMount{{
		Name:      "pach-bin",
		MountPath: "/pach-bin",
	}, {
		Name:      "pachyderm-worker",
		MountPath: "/pfs",
	}}

	storageVolumeMounts := []corev1.VolumeMount{{
		Name:      "pach-disk",
		MountPath: "/pach",
	}} //Missing storage secret

	userVolumeMounts := []corev1.VolumeMount{{
		Name:      "pach-bin",
		MountPath: "/pach-bin",
	}, {
		Name:      "pachyderm-worker",
		MountPath: "/pfs",
	}, {
		Name:      "pach-disk",
		MountPath: "/pach",
	}} //Missing storage secret and docker

	labels := map[string]string{
		"app":        "pachd",
		"controller": pipeline.Name,
	}
	zeroVal := int64(0)
	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("pipeline-%s-v1", pipeline.Name), //TODO Dry up with reconcile loop Deployment name
			Namespace: pipeline.Namespace,
			/*OwnerReferences: []metav1.OwnerReference{ //Added using SetControllerReference above
				*metav1.NewControllerRef(pipeline, ppsv1.SchemeGroupVersion.WithKind("Pipeline")),
			},*/
		},
		Spec: appsv1.DeploymentSpec{
			//Replicas: 1,
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels,
				},
				Spec: corev1.PodSpec{
					InitContainers: []corev1.Container{
						{
							Name:    "init",
							Image:   workerImage,
							Command: []string{"/app/init"},
							//Command: []string{"/pach/worker.sh"},
							//ImagePullPolicy: corev1.PullPolicy(pullPolicy),
							VolumeMounts: initVolumeMounts, //options.volumeMounts,
							/*Resources: corev1.ResourceRequirements{
								Requests: corev1.ResourceList{
									corev1.ResourceCPU:    cpuZeroQuantity,
									corev1.ResourceMemory: memDefaultQuantity,
								},
							},*/
						},
					},
					Containers: []corev1.Container{
						{
							Name:    "user",    //client.PPSWorkerUserContainerName,
							Image:   userImage, //options.userImage,
							Command: []string{"/pach-bin/worker"},
							/*
								Command: []string{"/pach-bin/dlv"},
								Args: []string{
									"exec", "/pach-bin/worker",
									"--listen=:2345",
									"--headless=true",
									//"--log=true",
									//"--log-output=debugger,debuglineerr,gdbwire,lldbout,rpc",
									"--accept-multiclient",
									"--api-version=2",
								},
							*/
							//ImagePullPolicy: v1.PullPolicy(pullPolicy),
							EnvFrom: envFrom,
							Env:     workerEnv,
							/*Resources: v1.ResourceRequirements{
								Requests: v1.ResourceList{
									v1.ResourceCPU:    cpuZeroQuantity,
									v1.ResourceMemory: memDefaultQuantity,
								},
							},*/
							VolumeMounts: userVolumeMounts,
						},
						{
							Name:  "storage", //client.PPSWorkerSidecarContainerName,
							Image: pachImage, //a.workerSidecarImage,
							//Command: []string{"/app/pachd", "--mode", "sidecar"},
							Command: []string{"/pachd", "--mode", "sidecar"},
							//ImagePullPolicy: v1.PullPolicy(pullPolicy),
							Env:          sidecarEnv,
							EnvFrom:      envFrom,
							VolumeMounts: storageVolumeMounts,
							/*Resources: v1.ResourceRequirements{
								Requests: v1.ResourceList{
									v1.ResourceCPU:    cpuZeroQuantity,
									v1.ResourceMemory: memSidecarQuantity,
								},
							},*/
							//Ports: sidecarPorts,
						},
					},
					SecurityContext:    &v1.PodSecurityContext{RunAsUser: &zeroVal},
					ServiceAccountName: "pachyderm-worker",
					//RestartPolicy: "Always",
					Volumes: []corev1.Volume{{
						Name: "pach-disk",
						VolumeSource: corev1.VolumeSource{
							HostPath: &corev1.HostPathVolumeSource{
								Path: "/var/pachyderm/pachd",
							},
						},
					}, {
						Name: "pachyderm-worker",
						VolumeSource: corev1.VolumeSource{
							EmptyDir: &corev1.EmptyDirVolumeSource{},
						},
					}, {
						Name: "pach-bin",
						VolumeSource: corev1.VolumeSource{
							EmptyDir: &corev1.EmptyDirVolumeSource{},
						},
					}},
					//ImagePullSecrets:              options.imagePullSecrets,
					//TerminationGracePeriodSeconds: &zeroVal,
					//SecurityContext:               securityContext,
				},
			},
		},
	}
}

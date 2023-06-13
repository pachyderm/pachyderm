package server

import (
	"context"
	"fmt"
	"path"
	"strconv"
	"strings"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"

	"github.com/pachyderm/pachyderm/v2/src/pps"

	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/errutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/limit"
	"github.com/pachyderm/pachyderm/v2/src/internal/log"
	"github.com/pachyderm/pachyderm/v2/src/internal/pachconfig"
	"github.com/pachyderm/pachyderm/v2/src/internal/ppsutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/tracing"
)

type kubeDriver struct {
	kubeClient kubernetes.Interface
	namespace  string
	// the limiter intends to guard the k8s API server from being overwhelmed by many concurrent requests
	// that could arise from many concurrent pipelineController goros.
	limiter    limit.ConcurrencyLimiter
	config     pachconfig.Configuration
	etcdPrefix string
}

func newKubeDriver(kubeClient kubernetes.Interface, config pachconfig.Configuration) InfraDriver {
	return &kubeDriver{
		kubeClient: kubeClient,
		namespace:  config.Namespace,
		limiter:    limit.New(config.PPSMaxConcurrentK8sRequests),
		config:     config,
		etcdPrefix: path.Join(config.EtcdPrefix, config.PPSEtcdPrefix),
	}
}

// Creates a pipeline's services, secrets, and replication controllers.
func (kd *kubeDriver) CreatePipelineResources(ctx context.Context, pi *pps.PipelineInfo) error {
	log.Info(ctx, "creating resources for pipeline")
	kd.limiter.Acquire()
	defer kd.limiter.Release()
	if err := kd.createWorkerSvcAndRc(ctx, pi); err != nil {
		if errors.As(err, &ppsutil.NoValidOptionsErr{}) {
			// these errors indicate invalid pipelineInfo, don't retry
			return stepError{
				error:        errors.Wrap(err, "could not generate RC options"),
				failPipeline: true,
			}
		}
		return newRetriableError(err, "error creating resources")
	}
	return nil
}

func pipelineLabelSelector(p *pps.Pipeline) string {
	selector := fmt.Sprintf("%s=%s", pipelineNameLabel, p.Name)
	if projectName := p.Project.GetName(); projectName != "" {
		selector = fmt.Sprintf("%s,%s=%s", selector, pipelineProjectLabel, projectName)
	} else {
		selector = fmt.Sprintf("%s,!%s", selector, pipelineProjectLabel)
	}
	return selector
}

// Deletes a pipeline's services, secrets, and replication controllers.
// NOTE: It doesn't return a stepError, leaving retry behavior to the caller
func (kd *kubeDriver) DeletePipelineResources(ctx context.Context, pipeline *pps.Pipeline) (retErr error) {
	projectName := pipeline.Project.GetName()
	pipelineName := pipeline.Name
	log.Info(ctx, "deleting resources for pipeline")
	span, ctx := tracing.AddSpanToAnyExisting(ctx,
		"/pps.Master/DeletePipelineResources", "project", projectName, "pipeline", pipelineName)
	defer func() {
		tracing.TagAnySpan(ctx, "err", retErr)
		tracing.FinishAnySpan(span)
	}()
	kd.limiter.Acquire()
	defer kd.limiter.Release()
	// Delete any services associated with pc.pipeline
	selector := pipelineLabelSelector(pipeline)
	opts := metav1.DeleteOptions{
		OrphanDependents: &falseVal,
	}
	services, err := kd.kubeClient.CoreV1().Services(kd.namespace).List(ctx, metav1.ListOptions{LabelSelector: selector})
	if err != nil {
		return errors.Wrapf(err, "could not list services")
	}
	for _, service := range services.Items {
		if err := kd.kubeClient.CoreV1().Services(kd.namespace).Delete(ctx, service.Name, opts); err != nil {
			if !errutil.IsNotFoundError(err) {
				return errors.Wrapf(err, "could not delete service %q", service.Name)
			}
		}
	}
	// Delete any secrets associated with pc.pipeline
	secrets, err := kd.kubeClient.CoreV1().Secrets(kd.namespace).List(ctx, metav1.ListOptions{LabelSelector: selector})
	if err != nil {
		return errors.Wrapf(err, "could not list secrets")
	}
	for _, secret := range secrets.Items {
		if err := kd.kubeClient.CoreV1().Secrets(kd.namespace).Delete(ctx, secret.Name, opts); err != nil {
			if !errutil.IsNotFoundError(err) {
				return errors.Wrapf(err, "could not delete secret %q", secret.Name)
			}
		}
	}
	// Finally, delete pc.pipeline's RC, which will cause pollPipelines to stop
	// polling it.
	rcs, err := kd.kubeClient.CoreV1().ReplicationControllers(kd.namespace).List(ctx, metav1.ListOptions{LabelSelector: selector})
	if err != nil {
		return errors.Wrapf(err, "could not list RCs")
	}
	for _, rc := range rcs.Items {
		if err := kd.kubeClient.CoreV1().ReplicationControllers(kd.namespace).Delete(ctx, rc.Name, opts); err != nil {
			if !errutil.IsNotFoundError(err) {
				return errors.Wrapf(err, "could not delete RC %q", rc.Name)
			}
		}
	}
	return nil
}

func (kd *kubeDriver) ReadReplicationController(ctx context.Context, pi *pps.PipelineInfo) (*v1.ReplicationControllerList, error) {
	kd.limiter.Acquire()
	defer kd.limiter.Release()
	// List all RCs, so stale RCs from old pipelines are noticed and deleted
	labelSelector := pipelineLabelSelector(pi.Pipeline)
	rc, err := kd.kubeClient.CoreV1().ReplicationControllers(kd.namespace).List(
		ctx,
		metav1.ListOptions{LabelSelector: labelSelector})
	return rc, errors.Wrapf(err, "failed to read rc for pipeline %s", pi.Pipeline)
}

// UpdateReplicationController intends to server {scaleUp,scaleDown}Pipeline.
// It includes all of the logic for writing an updated RC spec to kubernetes,
// and updating/retrying if k8s rejects the write. It presents a strange API,
// since the the RC being updated is already available to the caller, but update()
// may be called muliple times if the k8s write fails. It may be helpful to think
// of the 'old' rc passed to update() as mutable.
func (kd *kubeDriver) UpdateReplicationController(ctx context.Context, old *v1.ReplicationController, update func(rc *v1.ReplicationController) bool) error {
	// Apply op's update to rc
	rc := old.DeepCopy()
	// clearing resourceVersion effectively tells k8s to ignore concurrent modifications.
	// While this has a chance of overwriting a direct change to the RC outside of pachyderm,
	// the most likely source of conflict is replica state changes, which can be safely ignored.
	// Ensuring an uninterrupted update would require changing our retry logic,
	// as currently a failure here leads to failing the pipeline
	rc.ResourceVersion = ""
	if update(rc) {
		// write updated RC to k8s
		kd.limiter.Acquire()
		defer kd.limiter.Release()
		if _, err := kd.kubeClient.CoreV1().ReplicationControllers(kd.namespace).Update(
			ctx, rc, metav1.UpdateOptions{}); err != nil {
			return newRetriableError(err, "error updating RC")
		}
	}
	return nil
}

// Used to nudge pipeline controllers with refresh events.
func (kd *kubeDriver) ListReplicationControllers(ctx context.Context) (*v1.ReplicationControllerList, error) {
	rc, err := kd.kubeClient.CoreV1().ReplicationControllers(kd.namespace).List(
		ctx,
		metav1.ListOptions{
			LabelSelector: "suite=pachyderm,pipelineName",
		})
	return rc, errors.Wrap(err, "failed to list rcs")
}

// Used to discover crashing pods which signals the controller to transition
// a pipeline to CRASHING
func (kd *kubeDriver) WatchPipelinePods(ctx context.Context) (<-chan watch.Event, func(), error) {
	kubePipelineWatch, err := kd.kubeClient.CoreV1().Pods(kd.namespace).Watch(
		ctx,
		metav1.ListOptions{
			LabelSelector: metav1.FormatLabelSelector(metav1.SetAsLabelSelector(
				map[string]string{
					"component": "worker",
				})),
			Watch: true,
		})
	if err != nil {
		return nil, nil, errors.Wrap(err, "failed to watch kubernetes pods")
	}
	return kubePipelineWatch.ResultChan(), kubePipelineWatch.Stop, nil
}

// The following implement the ppsutil.K8sEnv interface.
func (kd *kubeDriver) DefaultCPURequest() resource.Quantity {
	return kd.config.PipelineDefaultCPURequest
}
func (kd *kubeDriver) DefaultMemoryRequest() resource.Quantity {
	return kd.config.PipelineDefaultMemoryRequest
}
func (kd *kubeDriver) DefaultStorageRequest() resource.Quantity {
	return kd.config.PipelineDefaultStorageRequest
}
func (kd *kubeDriver) ImagePullSecrets() []string {
	return strings.Split(kd.config.ImagePullSecrets, ",")
}
func (kd *kubeDriver) S3GatewayPort() uint16 {
	return kd.config.S3GatewayPort
}
func (kd *kubeDriver) PostgresSecretRef(ctx context.Context) (*v1.SecretKeySelector, error) {
	// Get the reference to the postgres secret used by the current pod
	podName := kd.config.PachdPodName
	selfPodInfo, err := kd.kubeClient.CoreV1().Pods(kd.namespace).Get(ctx, podName, metav1.GetOptions{})
	if err != nil {
		return nil, errors.EnsureStack(err)
	}
	var postgresSecretRef *v1.SecretKeySelector
	for _, container := range selfPodInfo.Spec.Containers {
		for _, envVar := range container.Env {
			if envVar.Name == "POSTGRES_PASSWORD" && envVar.ValueFrom != nil && envVar.ValueFrom.SecretKeyRef != nil {
				postgresSecretRef = envVar.ValueFrom.SecretKeyRef
			}
		}
	}
	if postgresSecretRef == nil {
		return nil, errors.New("could not load the existing postgres secret reference from kubernetes")
	}
	return postgresSecretRef, nil
}
func (kd *kubeDriver) ImagePullPolicy() string {
	return kd.config.WorkerImagePullPolicy
}
func (kd *kubeDriver) WorkerImage() string {
	return kd.config.WorkerImage
}
func (kd *kubeDriver) SidecarImage() string {
	return kd.config.WorkerSidecarImage
}
func (kd *kubeDriver) StorageRoot() string {
	return kd.config.StorageRoot
}
func (kd *kubeDriver) Namespace() string        { return kd.namespace }
func (kd *kubeDriver) StorageBackend() string   { return kd.config.StorageBackend }
func (kd *kubeDriver) PostgresUser() string     { return kd.config.PostgresUser }
func (kd *kubeDriver) PostgresDatabase() string { return kd.config.PostgresDBName }
func (kd *kubeDriver) PGBouncerHost() string    { return kd.config.PGBouncerHost }
func (kd *kubeDriver) PGBouncerPort() uint16    { return uint16(kd.config.PGBouncerPort) }
func (kd *kubeDriver) PeerPort() uint16         { return kd.config.PeerPort }
func (kd *kubeDriver) LokiHost() string         { return kd.config.LokiHost }
func (kd *kubeDriver) LokiPort() (uint16, error) {
	i, err := strconv.ParseUint(kd.config.LokiPort, 10, 16)
	if err != nil {
		return 0, err
	}
	return uint16(i), nil
}
func (kd *kubeDriver) SidecarPort() uint16           { return kd.config.Port }
func (kd *kubeDriver) InSidecars() bool              { return kd.config.PachwInSidecars }
func (kd *kubeDriver) GarbageCollectionPercent() int { return kd.config.GCPercent }
func (kd *kubeDriver) SidecarEnvVars(pipelineInfo *pps.PipelineInfo, commonEnv []v1.EnvVar) []v1.EnvVar {
	return append(commonEnv, kd.getEgressSecretEnvVars(pipelineInfo)...)
}
func (kd *kubeDriver) EtcdPrefix() string    { return kd.etcdPrefix }
func (kd *kubeDriver) PPSWorkerPort() uint16 { return kd.config.PPSWorkerPort }
func (kd *kubeDriver) CommitProgressCounterDisabled() bool {
	return kd.config.DisableCommitProgressCounter
}
func (kd *kubeDriver) LokiLoggingEnabled() bool { return kd.config.LokiLogging }
func (kd *kubeDriver) GoogleCloudProfilerProject() string {
	return kd.config.GoogleCloudProfilerProject
}
func (kd *kubeDriver) StorageHostPath() string { return kd.config.StorageHostPath }
func (kd *kubeDriver) TLSSecretName() string   { return kd.config.TLSCertSecretName }
func (kd *kubeDriver) WorkerSecurityContextsEnabled() bool {
	return kd.config.EnableWorkerSecurityContexts
}
func (kd *kubeDriver) WorkerUsesRoot() bool { return kd.config.WorkerUsesRoot }

package server

import (
	"context"
	"fmt"
	"path"

	"github.com/pachyderm/pachyderm/v2/src/client/limit"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/errutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/serviceenv"
	"github.com/pachyderm/pachyderm/v2/src/internal/tracing"
	"github.com/pachyderm/pachyderm/v2/src/pps"
	log "github.com/sirupsen/logrus"
	logrus "github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
)

type kubeDriver struct {
	kubeClient kubernetes.Interface
	namespace  string
	// the limiter intends to guard the k8s API server from being overwhelmed by many concurrent requests
	// that could arise from many concurrent pipelineController goros.
	limiter    limit.ConcurrencyLimiter
	config     serviceenv.Configuration
	etcdPrefix string
	logger     *logrus.Logger
}

func newKubeDriver(kubeClient kubernetes.Interface, config serviceenv.Configuration, logger *logrus.Logger) InfraDriver {
	return &kubeDriver{
		kubeClient: kubeClient,
		namespace:  config.Namespace,
		limiter:    limit.New(config.PPSMaxConcurrentK8sRequests),
		config:     config,
		etcdPrefix: path.Join(config.EtcdPrefix, config.PPSEtcdPrefix),
		logger:     logger,
	}
}

// Creates a pipeline's services, secrets, and replication controllers.
func (kd *kubeDriver) CreatePipelineResources(ctx context.Context, pi *pps.PipelineInfo) error {
	log.Infof("PPS master: creating resources for pipeline %q", pi.Pipeline.Name)
	kd.limiter.Acquire()
	defer kd.limiter.Release()
	if err := kd.createWorkerSvcAndRc(ctx, pi); err != nil {
		if errors.As(err, &noValidOptionsErr{}) {
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

// Deletes a pipeline's services, secrets, and replication controllers.
// NOTE: It doesn't return a stepError, leaving retry behavior to the caller
func (kd *kubeDriver) DeletePipelineResources(ctx context.Context, pipeline string) (retErr error) {
	log.Infof("PPS master: deleting resources for pipeline %q", pipeline)
	span, ctx := tracing.AddSpanToAnyExisting(ctx,
		"/pps.Master/DeletePipelineResources", "pipeline", pipeline)
	defer func() {
		tracing.TagAnySpan(ctx, "err", retErr)
		tracing.FinishAnySpan(span)
	}()
	kd.limiter.Acquire()
	defer kd.limiter.Release()
	// Delete any services associated with pc.pipeline
	selector := fmt.Sprintf("%s=%s", pipelineNameLabel, pipeline)
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
	rc, err := kd.kubeClient.CoreV1().ReplicationControllers(kd.namespace).List(
		ctx,
		metav1.ListOptions{LabelSelector: fmt.Sprintf("%s=%s", pipelineNameLabel, pi.Pipeline.Name)})
	return rc, errors.Wrapf(err, "failed to read rc for pipeline %s", pi.Pipeline.Name)
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

package server

import (
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
	"golang.org/x/net/context"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type infraState struct{}
type pipelineState struct{}

type kubeDriver struct {
	kubeClient *kubernetes.Clientset
	namespace  string
	// the limiter intends to guard the k8s API server from being overwhelmed by many concurrent requests
	// that could arise from many concurrent pipelineController goros.
	limiter    limit.ConcurrencyLimiter
	config     serviceenv.Configuration
	etcdPrefix string
	logger     *logrus.Logger
}

func newKubeDriver(kubeClient *kubernetes.Clientset, config serviceenv.Configuration, logger *logrus.Logger) *kubeDriver {
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

func (kd *kubeDriver) ReadReplicationController(ctx context.Context, pi *pps.PipelineInfo) (*v1.ReplicationController, error) {
	kd.limiter.Acquire()
	defer kd.limiter.Release()
	// List all RCs, so stale RCs from old pipelines are noticed and deleted
	rcs, err := kd.kubeClient.CoreV1().ReplicationControllers(kd.namespace).List(
		ctx,
		metav1.ListOptions{LabelSelector: fmt.Sprintf("%s=%s", pipelineNameLabel, pi.Pipeline.Name)})
	if err != nil && !errutil.IsNotFoundError(err) {
		return nil, errors.EnsureStack(err)
	}
	if len(rcs.Items) == 0 {
		return nil, errRCNotFound
	}
	rc := &rcs.Items[0]
	switch {
	case len(rcs.Items) > 1:
		// select stale RC if possible, so that we delete it in restartPipeline
		for i := range rcs.Items {
			rc = &rcs.Items[i]
			if !rcIsFresh(pi, rc) {
				break
			}
		}
		return nil, errTooManyRCs
	case !rcIsFresh(pi, rc):
		return nil, errStaleRC
	default:
		return rc, nil
	}
}

// UpdateReplicationController intends to server {scaleUp,scaleDown}Pipeline.
// It includes all of the logic for writing an updated RC spec to kubernetes,
// and updating/retrying if k8s rejects the write. It presents a strange API,
// since the the RC being updated is already available to the caller, but update()
// may be called muliple times if the k8s write fails. It may be helpful to think
// of the 'old' rc passed to update() as mutable.
func (kd *kubeDriver) UpdateReplicationController(ctx context.Context, old *v1.ReplicationController, update func(rc *v1.ReplicationController) bool) error {
	rcs := kd.kubeClient.CoreV1().ReplicationControllers(kd.namespace)
	kd.limiter.Acquire()
	defer kd.limiter.Release()
	// Apply op's update to rc
	rc := old.DeepCopy()
	if update(rc) {
		// write updated RC to k8s
		if _, err := rcs.Update(ctx, rc, metav1.UpdateOptions{}); err != nil {
			return newRetriableError(err, "error updating RC")
		}
	}
	return nil
}

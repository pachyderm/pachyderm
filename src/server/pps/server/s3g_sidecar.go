package server

import (
	"context"
	"fmt"
	"net/http"
	"path"
	"strings"
	"time"

	"go.uber.org/zap"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/pachyderm/pachyderm/v2/src/internal/client"
	"github.com/pachyderm/pachyderm/v2/src/pfs"
	"github.com/pachyderm/pachyderm/v2/src/pps"
	"github.com/pachyderm/pachyderm/v2/src/server/pfs/s3"

	"github.com/pachyderm/pachyderm/v2/src/internal/backoff"
	col "github.com/pachyderm/pachyderm/v2/src/internal/collection"
	"github.com/pachyderm/pachyderm/v2/src/internal/dlock"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/errutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/log"
	"github.com/pachyderm/pachyderm/v2/src/internal/ppsdb"
	"github.com/pachyderm/pachyderm/v2/src/internal/ppsutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/watch"
)

const (
	s3gSidecarLockPath = "_s3g_sidecar_lock"
)

type sidecarS3G struct {
	apiServer    *apiServer
	pipelineInfo *pps.PipelineInfo
	pachClient   *client.APIClient

	server *s3.S3Server
}

func (a *apiServer) ServeSidecarS3G(ctx context.Context) {
	s := &sidecarS3G{
		apiServer:  a,
		pachClient: a.env.GetPachClient(context.TODO()),
	}
	port := a.env.Config.S3GatewayPort
	s.server = s3.Server(ctx, port, nil)

	// Read spec commit for this sidecar's pipeline, and set auth token for pach
	// client
	specCommit := a.env.Config.PPSSpecCommitID
	if specCommit == "" {
		// This error is not recoverable
		panic("cannot serve sidecar S3 gateway if no spec commit is set")
	}
	if err := backoff.RetryNotify(func() error {
		var err error
		s.pipelineInfo, err = ppsutil.GetWorkerPipelineInfo(
			s.pachClient,
			a.env.DB,
			a.env.Listener,
			&pps.Pipeline{Project: &pfs.Project{Name: a.env.Config.PPSProjectName},
				Name: a.env.Config.PPSPipelineName},
			a.env.Config.PPSSpecCommitID,
		)
		return errors.Wrapf(err, "sidecar s3 gateway: could not find pipeline")
	}, backoff.NewInfiniteBackOff(), func(err error, d time.Duration) error {
		log.Error(ctx, "error starting sidecar s3 gateway; retrying", zap.Error(err), zap.Duration("retryAfter", d))
		return nil
	}); err != nil {
		// This code should never run, but I hesitate to introduce a panic to new
		// code
		log.Error(ctx, "restarting startup of sidecar s3 gateway", zap.Error(err))
		a.ServeSidecarS3G(ctx)
	}
	if !ppsutil.ContainsS3Inputs(s.pipelineInfo.Details.Input) && !s.pipelineInfo.Details.S3Out {
		return // break early (nothing to serve via S3 gateway)
	}

	go func() {
		for i := 0; i < 2; i++ { // If too many errors, the worker will fail the job
			err := s.server.ListenAndServe()
			if err == nil || errors.Is(err, http.ErrServerClosed) {
				break // server was shutdown/closed
			}
			log.Error(ctx, "error serving sidecar s3 gateway", zap.Error(err), log.RetryAttempt(i, 2))
		}
	}()

	// begin creating k8s services and s3 gateway instances for each job
	done := make(chan string)
	go func() {
		s.createK8sServices(ctx)
		done <- "createK8sServices"
	}()
	go func() {
		s.serveS3Instances(ctx)
		done <- "serveS3Instances"
	}()
	finisher := <-done
	log.Error(ctx, "process exited; this should never happen", zap.String("process", finisher))
	panic(
		fmt.Sprintf("sidecar s3 gateway: %s is exiting, which should never happen", finisher),
	)
}

type jobHandler interface {
	// OnCreate runs when a job is created. Should be idempotent.
	OnCreate(ctx context.Context, jobInfo *pps.JobInfo)

	// OnTerminate runs when a job ends. Should be idempotent.
	OnTerminate(ctx context.Context, job *pps.Job)
}

func (s *sidecarS3G) serveS3Instances(ctx context.Context) {
	// Watch for new jobs & initialize s3g for each new job
	(&handleJobsCtx{
		ctx: ctx,
		s:   s,
		h:   &s3InstanceCreatingJobHandler{s},
	}).start()
}

func (s *sidecarS3G) createK8sServices(ctx context.Context) {
	log.Info(ctx, "Launching sidecar s3 gateway master process")
	// createK8sServices goes through master election so that only one k8s service
	// is created per pachyderm job running sidecar s3 gateway
	var projectName = s.pipelineInfo.Pipeline.Project.GetName()
	backoff.RetryNotify(func() error { //nolint:errcheck
		masterLock := dlock.NewDLock(s.apiServer.env.EtcdClient,
			path.Join(s.apiServer.etcdPrefix,
				s3gSidecarLockPath,
				projectName,
				s.pipelineInfo.Pipeline.Name,
				s.pipelineInfo.Details.Salt))
		ctx, err := masterLock.Lock(s.pachClient.Ctx())
		if err != nil {
			// retry obtaining lock
			return errors.Wrapf(err, "error obtaining mastership")
		}

		// Watch for new jobs & create kubernetes service for each new job
		(&handleJobsCtx{
			ctx: ctx,
			s:   s,
			h:   &k8sServiceCreatingJobHandler{s},
		}).start()

		// Retry the unlock inside the larger retry as other sidecars may not be
		// able to obtain mastership until the key expires if unlock is unsuccessful
		if err := backoff.RetryNotify(func() error {
			return errors.EnsureStack(masterLock.Unlock(ctx))
		}, backoff.NewInfiniteBackOff(), func(err error, d time.Duration) error {
			log.Error(ctx, "Error releasing sidecar s3 gateway master lock; retrying", zap.Error(err), zap.Duration("retryAfter", d))
			return nil // always retry
		}); err != nil {
			return errors.Wrapf(err, "permanent error releasing sidecar s3 gateway master lock")
		}
		return nil
	}, backoff.NewInfiniteBackOff(), func(err error, d time.Duration) error {
		log.Error(ctx, "sidecar s3 gateway failed; retrying", zap.Error(err), zap.Duration("retryAfter", d))
		return nil
	})
}

type s3InstanceCreatingJobHandler struct {
	s *sidecarS3G
}

func (s *s3InstanceCreatingJobHandler) OnCreate(ctx context.Context, jobInfo *pps.JobInfo) {
	// serve new S3 gateway & add to s.server routers
	if ok := s.s.server.ContainsRouter(ppsutil.SidecarS3GatewayService(jobInfo.Job.Pipeline, jobInfo.Job.Id)); ok {
		return // s3g handler already created
	}

	// Initialize new S3 gateway
	var inputBuckets []*s3.Bucket
	err := pps.VisitInput(jobInfo.Details.Input, func(in *pps.Input) error {
		if in.Pfs != nil && in.Pfs.S3 {
			inputBuckets = append(inputBuckets, &s3.Bucket{
				Commit: client.NewSystemRepo(in.Pfs.Project, in.Pfs.Repo, in.Pfs.RepoType).NewCommit(in.Pfs.Branch, in.Pfs.Commit),
				Name:   in.Pfs.Name,
			})
		}
		return nil
	})
	log.Error(ctx, "could not visit pps input", zap.Error(err))
	var outputBucket *s3.Bucket
	if s.s.pipelineInfo.Details.S3Out {
		outputBucket = &s3.Bucket{
			Commit: jobInfo.OutputCommit,
			Name:   "out",
		}
	}
	driver := s3.NewWorkerDriver(inputBuckets, outputBucket)
	router := s3.Router(ctx, driver, s.s.apiServer.env.GetPachClient)
	s.s.server.AddRouter(ppsutil.SidecarS3GatewayService(jobInfo.Job.Pipeline, jobInfo.Job.Id), router)
}

func (s *s3InstanceCreatingJobHandler) OnTerminate(jobCtx context.Context, job *pps.Job) {
	s.s.server.RemoveRouter(ppsutil.SidecarS3GatewayService(job.Pipeline, job.Id))
}

type k8sServiceCreatingJobHandler struct {
	s *sidecarS3G
}

func (s *k8sServiceCreatingJobHandler) S3G() *sidecarS3G {
	return s.s
}

func (s *k8sServiceCreatingJobHandler) OnCreate(ctx context.Context, jobInfo *pps.JobInfo) {
	// Create kubernetes service for the current job ('jobInfo')
	copyMap := func(m map[string]string) map[string]string {
		nm := make(map[string]string)
		for k, v := range m {
			nm[k] = v
		}
		return nm
	}
	selectorlabels := map[string]string{
		// NOTE: this used to be ppsutil.PipelineRcName(jobInfo.Job.Pipeline.Name, jobInfo.PipelineVersion)
		appLabel:             "pipeline",
		pipelineNameLabel:    jobInfo.Job.Pipeline.Name,
		pipelineVersionLabel: fmt.Sprint(jobInfo.PipelineVersion),
		"suite":              "pachyderm",
		"component":          "worker",
	}
	if projectName := jobInfo.Job.Pipeline.Project.GetName(); projectName != "" {
		selectorlabels[pipelineProjectLabel] = projectName
	}
	svcLabels := copyMap(selectorlabels)
	svcLabels["job"] = jobInfo.Job.Id // for reference, we also want to leave info about the job in the service definition
	service := &v1.Service{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Service",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:   ppsutil.SidecarS3GatewayService(jobInfo.Job.Pipeline, jobInfo.Job.Id),
			Labels: svcLabels,
		},
		Spec: v1.ServiceSpec{
			Selector: selectorlabels,
			// Create a headless service so that the worker's kube proxy doesn't
			// have to get a routing path for the service IP (i.e. the worker kube
			// proxy can have stale routes and clients running inside the worker
			// can still connect)
			ClusterIP: "None",
			Ports: []v1.ServicePort{
				{
					Port: int32(s.s.apiServer.env.Config.S3GatewayPort),
					Name: "s3-gateway-port",
				},
			},
		},
	}

	err := backoff.RetryNotify(func() error {
		_, err := s.s.apiServer.env.KubeClient.CoreV1().Services(s.s.apiServer.namespace).Create(ctx, service, metav1.CreateOptions{})
		if err != nil && strings.Contains(err.Error(), "already exists") {
			return nil // service already created
		}
		return errors.EnsureStack(err)
	}, backoff.NewExponentialBackOff(), func(err error, d time.Duration) error {
		log.Error(ctx, "error creating kubernetes service for s3 gateway sidecar; retrying", zap.Error(err), zap.Duration("retryAfter", d))
		return nil
	})
	if err != nil {
		log.Error(ctx, "could not create service for job", zap.Stringer("job", jobInfo.Job), zap.Error(err))
	}
}

func (s *k8sServiceCreatingJobHandler) OnTerminate(ctx context.Context, job *pps.Job) {
	if !ppsutil.ContainsS3Inputs(s.s.pipelineInfo.Details.Input) && !s.s.pipelineInfo.Details.S3Out {
		return // Nothing to delete; this isn't an s3 pipeline (shouldn't happen)
	}
	if err := backoff.RetryNotify(func() error {
		err := s.s.apiServer.env.KubeClient.CoreV1().Services(s.s.apiServer.namespace).Delete(
			ctx,
			ppsutil.SidecarS3GatewayService(job.Pipeline, job.Id),
			metav1.DeleteOptions{OrphanDependents: new(bool) /* false */})
		if err != nil && errutil.IsNotFoundError(err) {
			return nil // service already deleted
		}
		return errors.EnsureStack(err)
	}, backoff.NewExponentialBackOff(), func(err error, d time.Duration) error {
		log.Info(ctx, "error deleting kubernetes service for s3 gateway sidecar; retrying", log.Proto("job", job), zap.Error(err), zap.Duration("retryAfter", d))
		return nil
	}); err != nil {
		log.Error(ctx, "permanent error deleting kubernetes service", log.Proto("job", job), zap.Error(err))
	}
}

type handleJobsCtx struct {
	ctx context.Context
	s   *sidecarS3G
	h   jobHandler
}

func (h *handleJobsCtx) start() {
	defer func() {
		panic("sidecar s3 gateway: start() is exiting; this should never happen")
	}()
	for { // reestablish watch in a loop, in case there's a watch error
		var watcher watch.Watcher
		backoff.Retry(func() error { //nolint:errcheck
			var err error
			watcher, err = h.s.apiServer.jobs.ReadOnly(context.Background()).WatchByIndex(
				ppsdb.JobsPipelineIndex, ppsdb.JobsPipelineKey(h.s.pipelineInfo.Pipeline))
			if err != nil {
				return errors.Wrapf(err, "error creating watch")
			}
			return nil
		}, backoff.NewInfiniteBackOff())

		for e := range watcher.Watch() {
			if e.Type == watch.EventError {
				log.Error(h.ctx, "sidecar s3 gateway watch error", zap.Error(e.Err))
				break // reestablish watch
			}

			var key string
			jobInfo := &pps.JobInfo{}
			if err := e.Unmarshal(&key, jobInfo); err != nil {
				log.Error(h.ctx, "sidecar s3 gateway watch unmarshal error", zap.Error(err))
			}

			h.processJobEvent(context.Background(), e.Type, jobInfo.Job)
		}
		watcher.Close()
	}
}

func (h *handleJobsCtx) processJobEvent(jobCtx context.Context, t watch.EventType, job *pps.Job) {
	if t == watch.EventDelete {
		h.h.OnTerminate(jobCtx, job)
		return
	}
	// 'e' is a Put event (new or updated job)
	pachClient := h.s.pachClient.WithCtx(jobCtx)
	// Inspect the job and make sure it's relevant, as this worker may be old
	log.Info(h.ctx, "sidecar s3 gateway: inspecting job to begin serving inputs over s3 gateway", log.Proto("job", job))

	var jobInfo *pps.JobInfo
	if err := backoff.RetryNotify(func() error {
		var err error
		jobInfo, err = pachClient.InspectJob(h.s.pipelineInfo.Pipeline.Project.GetName(), h.s.pipelineInfo.Pipeline.Name, job.Id, true)
		if err != nil {
			if col.IsErrNotFound(err) {
				// TODO(msteffen): I'm not sure what this means--maybe that the service
				// was created and immediately deleted, and there's a pending deletion
				// event? In any case, without a job that exists there's nothing to act on
				log.Error(h.ctx, "sidecar s3 gateway: job not found", log.Proto("job", job), zap.Error(err))
				return nil
			}
			return err
		}
		return nil
	}, backoff.NewExponentialBackOff(), func(err error, d time.Duration) error {
		log.Error(h.ctx, "error inspecting job; retrying", log.Proto("job", job), zap.Error(err), zap.Duration("retryAfter", d))
		return nil
	}); err != nil {
		log.Error(h.ctx, "permanent error inspecting job", log.Proto("job", job), zap.Error(err))
		return // leak the job; better than getting stuck?
	}
	if jobInfo.PipelineVersion < h.s.pipelineInfo.Version {
		log.Info(h.ctx, "skipping job as it uses old pipeline version", log.Proto("job", job), zap.Uint64("jobVersion", jobInfo.PipelineVersion), zap.Uint64("ourVersion", h.s.pipelineInfo.Version))
		return
	}
	if jobInfo.PipelineVersion > h.s.pipelineInfo.Version {
		log.Info(h.ctx, "skipping job as its pipeline version version is "+
			"greater than this worker's pipeline version; this should "+
			"automatically resolve when the worker is updated",
			log.Proto("job", job), zap.Uint64("jobVersion", jobInfo.PipelineVersion), zap.Uint64("ourVersion", h.s.pipelineInfo.Version))
		return
	}
	if pps.IsTerminal(jobInfo.State) {
		h.h.OnTerminate(jobCtx, job)
		return
	}

	h.h.OnCreate(jobCtx, jobInfo)
}

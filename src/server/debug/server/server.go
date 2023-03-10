package server

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"os"
	"path"
	"reflect"
	"runtime"
	runtimedebug "runtime/debug"
	"runtime/pprof"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/wcharczuk/go-chart"
	"go.uber.org/multierr"
	"go.uber.org/zap"
	"golang.org/x/exp/maps"
	"gopkg.in/yaml.v3"
	v1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	describe "k8s.io/kubectl/pkg/describe"

	"github.com/gogo/protobuf/jsonpb"
	"github.com/gogo/protobuf/types"

	"github.com/pachyderm/pachyderm/v2/src/client"
	"github.com/pachyderm/pachyderm/v2/src/debug"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/errutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/grpcutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/log"
	loki "github.com/pachyderm/pachyderm/v2/src/internal/lokiutil/client"
	"github.com/pachyderm/pachyderm/v2/src/internal/pachsql"
	"github.com/pachyderm/pachyderm/v2/src/internal/ppsutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/serviceenv"
	"github.com/pachyderm/pachyderm/v2/src/pfs"
	"github.com/pachyderm/pachyderm/v2/src/pps"
	workerserver "github.com/pachyderm/pachyderm/v2/src/server/worker/server"
)

const (
	defaultDuration = time.Minute
	pachdPrefix     = "pachd"
	pipelinePrefix  = "pipelines"
	podPrefix       = "pods"
	databasePrefix  = "database"
)

type debugServer struct {
	env           serviceenv.ServiceEnv
	name          string
	sidecarClient *client.APIClient
	marshaller    *jsonpb.Marshaler
	database      *pachsql.DB

	logLevel, grpcLevel log.LevelChanger
}

// NewDebugServer creates a new server that serves the debug api over GRPC
func NewDebugServer(env serviceenv.ServiceEnv, name string, sidecarClient *client.APIClient, db *pachsql.DB) debug.DebugServer {
	return &debugServer{
		env:           env,
		name:          name,
		sidecarClient: sidecarClient,
		marshaller:    &jsonpb.Marshaler{Indent: "  "},
		database:      db,
		logLevel:      log.LogLevel,
		grpcLevel:     log.GRPCLevel,
	}
}

type collectPipelineFunc func(context.Context, *tar.Writer, *client.APIClient, *pps.PipelineInfo, ...string) error
type collectWorkerFunc func(context.Context, *tar.Writer, *v1.Pod, ...string) error
type redirectFunc func(context.Context, debug.DebugClient, *debug.Filter) (io.Reader, error)
type collectFunc func(context.Context, *tar.Writer, *client.APIClient, ...string) error

func (s *debugServer) handleRedirect(
	pachClient *client.APIClient,
	server grpcutil.StreamingBytesServer,
	filter *debug.Filter,
	collectPachd collectFunc,
	collectPipeline collectPipelineFunc,
	collectWorker collectWorkerFunc,
	redirect redirectFunc,
	collect collectFunc,
	wantAppLogs bool,
	wantDatabase bool,
) error {
	ctx := pachClient.Ctx() // this context has authorization credentials we need
	return grpcutil.WithStreamingBytesWriter(server, func(w io.Writer) error {
		return withDebugWriter(w, func(tw *tar.Writer) error {
			// Handle filter.
			pachdContainerPrefix := join(pachdPrefix, s.name, "pachd")
			if filter != nil {
				switch f := filter.Filter.(type) {
				case *debug.Filter_Pachd:
					var errs error
					if err := collectPachd(ctx, tw, pachClient, pachdContainerPrefix); err != nil {
						multierr.AppendInto(&errs, errors.Wrap(err, "collectPachd"))
					}
					if wantDatabase {
						if err := s.collectDatabaseDump(ctx, tw, databasePrefix); err != nil {
							multierr.AppendInto(&errs, errors.Wrap(err, "collectDatabaseDump"))
						}
					}
					return errs
				case *debug.Filter_Pipeline:
					var errs error
					pipelineInfo, err := pachClient.InspectProjectPipeline(f.Pipeline.Project.GetName(), f.Pipeline.Name, true)
					if err != nil {
						multierr.AppendInto(&errs, errors.Wrapf(err, "inspectProjectPipeline(%s)", f.Pipeline))
					}
					if err := s.handlePipelineRedirect(ctx, tw, pachClient, pipelineInfo, collectPipeline, collectWorker, redirect); err != nil {
						multierr.AppendInto(&errs, errors.Wrapf(err, "handlePipelineRedirect(%s)", pipelineInfo.GetPipeline()))
					}
					return errs
				case *debug.Filter_Worker:
					if f.Worker.Redirected {
						var errs error
						// Collect self if we're not a worker.
						if s.sidecarClient == nil {
							return collect(ctx, tw, pachClient, client.PPSWorkerSidecarContainerName)
						}
						// Collect the user container.
						if err := collect(ctx, tw, pachClient, client.PPSWorkerUserContainerName); err != nil {
							multierr.AppendInto(&errs, errors.Wrap(err, "collect user container"))
						}
						// Redirect to the storage container.
						r, err := redirect(ctx, s.sidecarClient.DebugClient, filter)
						if err != nil {
							multierr.AppendInto(&errs, errors.Wrap(err, "collect storage container"))
						}
						if err := collectDebugStream(tw, r); err != nil {
							multierr.AppendInto(&errs, errors.Wrap(err, "collect debug stream"))
						}
						return errs
					}
					pod, err := s.env.GetKubeClient().CoreV1().Pods(s.env.Config().Namespace).Get(ctx, f.Worker.Pod, metav1.GetOptions{})
					if err != nil {
						return errors.Wrapf(err, "getPod(%s)", f.Worker.Pod)
					}
					if err := s.handleWorkerRedirect(ctx, tw, pod, collectWorker, redirect); err != nil {
						return errors.Wrapf(err, "handleWorkerRedirect(%s)", f.Worker.Pod)
					}
					return nil
				case *debug.Filter_Database:
					if wantDatabase {
						return s.collectDatabaseDump(ctx, tw, databasePrefix)
					}
				}
			}

			// No filter, return everything
			var errs error
			if err := collectPachd(ctx, tw, pachClient, pachdContainerPrefix); err != nil {
				multierr.AppendInto(&errs, errors.Wrap(err, "collectPachd"))
			}
			pipelineInfos, err := pachClient.ListPipeline(true)
			if err != nil {
				multierr.AppendInto(&errs, errors.Wrap(err, "listPipelines"))
			}
			for _, pipelineInfo := range pipelineInfos {
				if err := s.handlePipelineRedirect(ctx, tw, pachClient, pipelineInfo, collectPipeline, collectWorker, redirect); err != nil {
					multierr.AppendInto(&errs, errors.Wrapf(err, "handlePipelineRedirect(%s)", pipelineInfo.GetPipeline()))
				}
			}
			if wantAppLogs {
				// All other pachyderm apps (console, pg-bouncer, etcd, pachw, etc.).
				if err := s.appLogs(ctx, tw); err != nil {
					multierr.AppendInto(&errs, errors.Wrap(err, "appLogs"))
				}
			}
			if wantDatabase {
				if err := s.collectDatabaseDump(ctx, tw, databasePrefix); err != nil {
					multierr.AppendInto(&errs, errors.Wrap(err, "collectDatabaseDump"))
				}
			}
			if err := s.helmReleases(ctx, tw); err != nil {
				multierr.AppendInto(&errs, errors.Wrap(err, "helmReleases"))
			}
			return errs
		})
	})
}

func (s *debugServer) appLogs(ctx context.Context, tw *tar.Writer) (retErr error) {
	ctx, end := log.SpanContext(ctx, "collectAppLogs")
	defer end(log.Errorp(&retErr))
	ctx, c := context.WithTimeout(ctx, 10*time.Minute)
	defer c()

	pods, err := s.env.GetKubeClient().CoreV1().Pods(s.env.Config().Namespace).List(ctx, metav1.ListOptions{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ListOptions",
			APIVersion: "v1",
		},
		LabelSelector: metav1.FormatLabelSelector(&metav1.LabelSelector{
			MatchLabels: map[string]string{
				"suite": "pachyderm",
			},
			MatchExpressions: []metav1.LabelSelectorRequirement{
				{
					Key:      "component",
					Operator: metav1.LabelSelectorOpNotIn,
					// Worker and pachd logs are collected by separate
					// functions.
					Values: []string{"worker", "pachd"},
				},
			},
		}),
	})
	if err != nil {
		return errors.EnsureStack(err)
	}
	var errs error
	for _, pod := range pods.Items {
		podPrefix := join(pod.Labels["app"], pod.Name)
		if err := s.collectDescribe(ctx, tw, pod.Name, podPrefix); err != nil {
			multierr.AppendInto(&errs, errors.Wrapf(err, "describe(%s)", pod.Name))
		}
		for _, container := range pod.Spec.Containers {
			prefix := join(podPrefix, container.Name)
			if err := s.collectLogs(ctx, tw, pod.Name, container.Name, prefix); err != nil {
				multierr.AppendInto(&errs, errors.Wrapf(err, "collectLogs(%s.%s)", pod.Name, container.Name))
			}
			if err := s.collectLogsLoki(ctx, tw, pod.Name, container.Name, prefix); err != nil {
				multierr.AppendInto(&errs, errors.Wrapf(err, "collectLogsLoki(%s.%s)", pod.Name, container.Name))
			}
		}
	}
	return errs
}

func (s *debugServer) handlePipelineRedirect(
	ctx context.Context,
	tw *tar.Writer,
	pachClient *client.APIClient,
	pipelineInfo *pps.PipelineInfo,
	collectPipeline collectPipelineFunc,
	collectWorker collectWorkerFunc,
	redirect redirectFunc,
) (retErr error) {
	if pipelineInfo == nil {
		return errors.Errorf("nil pipeline info")
	}
	if pipelineInfo.Pipeline == nil {
		return errors.Errorf("nil pipeline in pipeline info")
	}
	if pipelineInfo.Pipeline.Project == nil {
		return errors.Errorf("nil project in pipeline %q", pipelineInfo.Pipeline.Name)
	}
	prefix := join(pipelinePrefix, pipelineInfo.Pipeline.Project.Name, pipelineInfo.Pipeline.Name)
	defer func() {
		if retErr != nil {
			retErr = writeErrorFile(tw, retErr, prefix)
		}
	}()
	if collectPipeline != nil {
		if err := collectPipeline(ctx, tw, pachClient, pipelineInfo, prefix); err != nil {
			return err
		}
	}
	return s.forEachWorker(ctx, pipelineInfo, func(pod *v1.Pod) error {
		return s.handleWorkerRedirect(ctx, tw, pod, collectWorker, redirect, prefix)
	})
}

func (s *debugServer) forEachWorker(ctx context.Context, pipelineInfo *pps.PipelineInfo, cb func(*v1.Pod) error) error {
	pods, err := s.getWorkerPods(ctx, pipelineInfo)
	if err != nil {
		return err
	}
	if len(pods) == 0 {
		return errors.Errorf("no worker pods found for pipeline %s", pipelineInfo.GetPipeline())
	}
	var errs error
	for _, pod := range pods {
		if err := cb(&pod); err != nil {
			multierr.AppendInto(&errs, errors.Wrapf(err, "forEachWorker(%s)", pod.Name))
		}
	}
	return nil
}

func (s *debugServer) getWorkerPods(ctx context.Context, pipelineInfo *pps.PipelineInfo) ([]v1.Pod, error) {
	podList, err := s.env.GetKubeClient().CoreV1().Pods(s.env.Config().Namespace).List(
		ctx,
		metav1.ListOptions{
			TypeMeta: metav1.TypeMeta{
				Kind:       "ListOptions",
				APIVersion: "v1",
			},
			LabelSelector: metav1.FormatLabelSelector(
				metav1.SetAsLabelSelector(
					map[string]string{
						"app":             "pipeline",
						"pipelineProject": pipelineInfo.Pipeline.Project.Name,
						"pipelineName":    pipelineInfo.Pipeline.Name,
						"pipelineVersion": fmt.Sprint(pipelineInfo.Version),
					},
				),
			),
		},
	)
	if err != nil {
		if apierrors.IsNotFound(err) {
			return s.getLegacyWorkerPods(ctx, pipelineInfo)
		}
		return nil, errors.EnsureStack(err)
	}
	if len(podList.Items) == 0 {
		return s.getLegacyWorkerPods(ctx, pipelineInfo)
	}
	return podList.Items, nil
}

func (s *debugServer) getLegacyWorkerPods(ctx context.Context, pipelineInfo *pps.PipelineInfo) ([]v1.Pod, error) {
	podList, err := s.env.GetKubeClient().CoreV1().Pods(s.env.Config().Namespace).List(
		ctx,
		metav1.ListOptions{
			TypeMeta: metav1.TypeMeta{
				Kind:       "ListOptions",
				APIVersion: "v1",
			},
			LabelSelector: metav1.FormatLabelSelector(
				metav1.SetAsLabelSelector(
					map[string]string{
						"app": ppsutil.PipelineRcName(pipelineInfo),
					},
				),
			),
		},
	)
	if err != nil {
		return nil, errors.EnsureStack(err)
	}
	return podList.Items, nil
}

func (s *debugServer) handleWorkerRedirect(ctx context.Context, tw *tar.Writer, pod *v1.Pod, collectWorker collectWorkerFunc, cb redirectFunc, prefix ...string) (retErr error) {
	workerPrefix := join(podPrefix, pod.Name)
	if len(prefix) > 0 {
		workerPrefix = join(prefix[0], workerPrefix)
	}

	ctx, end := log.SpanContext(ctx, "handleWorkerRedirect", zap.String("pod", pod.Name))
	defer end(log.Errorp(&retErr))

	defer func() {
		if retErr != nil {
			retErr = writeErrorFile(tw, retErr, workerPrefix)
		}
	}()
	if collectWorker != nil {
		if err := collectWorker(ctx, tw, pod, workerPrefix); err != nil {
			return err
		}
	}
	if pod.Status.Phase != v1.PodRunning {
		return errors.Errorf("pod in phase %v, must be in phase %v to collect debug information", pod.Status.Phase, v1.PodRunning)
	}
	c, err := workerserver.NewClient(pod.Status.PodIP)
	if err != nil {
		return err
	}
	defer func() {
		if err := c.Close(); err != nil {
			log.Error(ctx, "errored closing worker client", zap.Error(err))
		}
	}()
	r, err := cb(ctx, c.DebugClient, &debug.Filter{
		Filter: &debug.Filter_Worker{
			Worker: &debug.Worker{
				Pod:        pod.Name,
				Redirected: true,
			},
		},
	})
	if err != nil {
		return err
	}
	return collectDebugStream(tw, r, workerPrefix)
}

func (s *debugServer) Profile(request *debug.ProfileRequest, server debug.Debug_ProfileServer) error {
	pachClient := s.env.GetPachClient(server.Context())
	return s.handleRedirect(
		pachClient,
		server,
		request.Filter,
		collectProfileFunc(request.Profile),  /* collectPachd */
		nil,                                  /* collectPipeline */
		nil,                                  /* collectWorker */
		redirectProfileFunc(request.Profile), /* redirect */
		collectProfileFunc(request.Profile),  /* collect */
		false,
		false,
	)
}

func collectProfileFunc(profile *debug.Profile) collectFunc {
	return func(ctx context.Context, tw *tar.Writer, _ *client.APIClient, prefix ...string) error {
		return collectProfile(ctx, tw, profile, prefix...)
	}
}

func collectProfile(ctx context.Context, tw *tar.Writer, profile *debug.Profile, prefix ...string) error {
	return collectDebugFile(tw, profile.Name, "", func(w io.Writer) error {
		return writeProfile(ctx, w, profile)
	}, prefix...)
}

func writeProfile(ctx context.Context, w io.Writer, profile *debug.Profile) (retErr error) {
	defer log.Span(ctx, "writeProfile", zap.String("profile", profile.GetName()))(log.Errorp(&retErr))
	if profile.Name == "cpu" {
		if err := pprof.StartCPUProfile(w); err != nil {
			return errors.EnsureStack(err)
		}
		defer pprof.StopCPUProfile()
		duration := defaultDuration
		if profile.Duration != nil {
			var err error
			duration, err = types.DurationFromProto(profile.Duration)
			if err != nil {
				return errors.EnsureStack(err)
			}
		}
		// Wait for either the defined duration, or until the context is
		// done.
		t := time.NewTimer(duration)
		select {
		case <-ctx.Done():
			t.Stop()
			return errors.EnsureStack(ctx.Err())
		case <-t.C:
			return nil
		}
	}
	p := pprof.Lookup(profile.Name)
	if p == nil {
		return errors.Errorf("unable to find profile %q", profile.Name)
	}
	return errors.EnsureStack(p.WriteTo(w, 0))
}

func redirectProfileFunc(profile *debug.Profile) redirectFunc {
	return func(ctx context.Context, c debug.DebugClient, filter *debug.Filter) (io.Reader, error) {
		profileC, err := c.Profile(ctx, &debug.ProfileRequest{
			Profile: profile,
			Filter:  filter,
		})
		if err != nil {
			return nil, errors.EnsureStack(err)
		}
		return grpcutil.NewStreamingBytesReader(profileC, nil), nil
	}
}

func (s *debugServer) Binary(request *debug.BinaryRequest, server debug.Debug_BinaryServer) error {
	pachClient := s.env.GetPachClient(server.Context())
	return s.handleRedirect(
		pachClient,
		server,
		request.Filter,
		collectBinary,  /* collectPachd */
		nil,            /* collectPipeline */
		nil,            /* collectWorker */
		redirectBinary, /* redirect */
		collectBinary,  /* collect */
		false,
		false,
	)
}

func collectBinary(ctx context.Context, tw *tar.Writer, _ *client.APIClient, prefix ...string) error {
	return collectDebugFile(tw, "binary", "", func(w io.Writer) (retErr error) {
		defer log.Span(ctx, "collectBinary", zap.String("binary", os.Args[0]))(log.Errorp(&retErr))
		f, err := os.Open(os.Args[0])
		if err != nil {
			return errors.EnsureStack(err)
		}
		defer func() {
			if err := f.Close(); retErr == nil {
				retErr = err
			}
		}()
		_, err = io.Copy(w, f)
		return errors.EnsureStack(err)
	}, prefix...)
}

func redirectBinary(ctx context.Context, c debug.DebugClient, filter *debug.Filter) (io.Reader, error) {
	binaryC, err := c.Binary(ctx, &debug.BinaryRequest{Filter: filter})
	if err != nil {
		return nil, errors.EnsureStack(err)
	}
	return grpcutil.NewStreamingBytesReader(binaryC, nil), nil
}

func (s *debugServer) Dump(request *debug.DumpRequest, server debug.Debug_DumpServer) error {
	if request.Limit == 0 {
		request.Limit = math.MaxInt64
	}
	timeout := 25 * time.Minute // If no client-side deadline set, use 25 minutes.
	if deadline, ok := server.Context().Deadline(); ok {
		d := time.Until(deadline)
		// Time out our own operations at ~90% of the client-set deadline. (22 minutes of
		// server-side processing for every 25 minutes of debug dumping.)
		timeout = time.Duration(22) * d / time.Duration(25)
	}
	ctx, c := context.WithTimeout(server.Context(), timeout)
	defer c()
	pachClient := s.env.GetPachClient(ctx)
	return s.handleRedirect(
		pachClient,
		server,
		request.Filter,
		s.collectPachdDumpFunc(request.Limit),    /* collectPachd */
		s.collectPipelineDumpFunc(request.Limit), /* collectPipeline */
		s.collectWorkerDump,                      /* collectWorker */
		redirectDump,                             /* redirect */
		collectDump,                              /* collect */
		true,
		true,
	)
}

func (s *debugServer) collectPachdDumpFunc(limit int64) collectFunc {
	return func(ctx context.Context, tw *tar.Writer, pachClient *client.APIClient, prefix ...string) (retErr error) {
		ctx, end := log.SpanContext(ctx, "collectPachdDump")
		defer end(log.Errorp(&retErr))

		var errs error
		// Collect input repos.
		if err := s.collectInputRepos(ctx, tw, pachClient, limit); err != nil {
			multierr.AppendInto(&errs, errors.Wrap(err, "collectInputRepos"))
		}
		// Collect the pachd version.
		if err := s.collectPachdVersion(ctx, tw, pachClient, prefix...); err != nil {
			multierr.AppendInto(&errs, errors.Wrap(err, "collectPachdVersion"))
		}
		// Collect the pachd describe output.
		if err := s.collectDescribe(ctx, tw, s.name, prefix...); err != nil {
			multierr.AppendInto(&errs, errors.Wrap(err, "collectDescribe"))
		}
		// Collect the pachd container logs.
		if err := s.collectLogs(ctx, tw, s.name, "pachd", prefix...); err != nil {
			multierr.AppendInto(&errs, errors.Wrap(err, "collectLogs"))
		}
		// Collect the pachd container logs from loki.
		lctx, c := context.WithTimeout(ctx, time.Minute)
		defer c()
		if err := s.collectLogsLoki(lctx, tw, s.name, "pachd", prefix...); err != nil {
			multierr.AppendInto(&errs, errors.Wrap(err, "collectLogsLoki"))
		}
		// Collect the pachd container dump.
		if err := collectDump(ctx, tw, pachClient, prefix...); err != nil {
			multierr.AppendInto(&errs, errors.Wrap(err, "collectDump"))
		}
		return errs
	}
}

func (s *debugServer) collectInputRepos(ctx context.Context, tw *tar.Writer, pachClient *client.APIClient, limit int64) (retErr error) {
	defer log.Span(ctx, "collectInputRepos")(log.Errorp(&retErr))
	repoInfos, err := pachClient.ListRepo()
	if err != nil {
		return err
	}
	var errs error
	for i, repoInfo := range repoInfos {
		if err := validateRepoInfo(repoInfo); err != nil {
			multierr.AppendInto(&errs, errors.Wrapf(err, "invalid repo info %d (%s) from ListRepo", i, repoInfo.String()))
			continue
		}
		if _, err := pachClient.InspectProjectPipeline(repoInfo.Repo.Project.Name, repoInfo.Repo.Name, true); err != nil {
			if errutil.IsNotFoundError(err) {
				repoPrefix := join("source-repos", repoInfo.Repo.Project.Name, repoInfo.Repo.Name)
				if err := s.collectCommits(ctx, tw, pachClient, repoInfo.Repo, limit, repoPrefix); err != nil {
					multierr.AppendInto(&errs, errors.Wrapf(err, "collectCommits(%s:%s)", repoInfo.GetRepo(), repoPrefix))
				}
				continue
			}
			multierr.AppendInto(&errs, errors.Wrapf(err, "inspectPipeline(%s)", repoInfo.GetRepo()))
		}
	}
	return errs
}

func (s *debugServer) collectCommits(rctx context.Context, tw *tar.Writer, pachClient *client.APIClient, repo *pfs.Repo, limit int64, prefix ...string) error {
	compacting := chart.ContinuousSeries{
		Name: "compacting",
		Style: chart.Style{
			Show:        true,
			StrokeColor: chart.GetDefaultColor(0).WithAlpha(255),
		},
	}
	validating := chart.ContinuousSeries{
		Name: "validating",
		Style: chart.Style{
			Show:        true,
			StrokeColor: chart.GetDefaultColor(1).WithAlpha(255),
		},
	}
	finishing := chart.ContinuousSeries{
		Name: "finishing",
		Style: chart.Style{
			Show:        true,
			StrokeColor: chart.GetDefaultColor(2).WithAlpha(255),
		},
	}
	if err := collectDebugFile(tw, "commits", "json", func(w io.Writer) error {
		ctx, cancel := context.WithCancel(rctx)
		defer cancel()
		client, err := pachClient.PfsAPIClient.ListCommit(ctx, &pfs.ListCommitRequest{
			Repo:   repo,
			Number: limit,
			All:    true,
		})
		if err != nil {
			return errors.EnsureStack(err)
		}
		return grpcutil.ForEach[*pfs.CommitInfo](client, func(ci *pfs.CommitInfo) error {
			if ci.Finished != nil && ci.Details != nil && ci.Details.CompactingTime != nil && ci.Details.ValidatingTime != nil {
				compactingDuration, err := types.DurationFromProto(ci.Details.CompactingTime)
				if err != nil {
					return errors.EnsureStack(err)
				}
				compacting.XValues = append(compacting.XValues, float64(len(compacting.XValues)+1))
				compacting.YValues = append(compacting.YValues, float64(compactingDuration))
				validatingDuration, err := types.DurationFromProto(ci.Details.ValidatingTime)
				if err != nil {
					return errors.EnsureStack(err)
				}
				validating.XValues = append(validating.XValues, float64(len(validating.XValues)+1))
				validating.YValues = append(validating.YValues, float64(validatingDuration))
				finishingTime, err := types.TimestampFromProto(ci.Finishing)
				if err != nil {
					return errors.EnsureStack(err)
				}
				finishedTime, err := types.TimestampFromProto(ci.Finished)
				if err != nil {
					return errors.EnsureStack(err)
				}
				finishingDuration := finishedTime.Sub(finishingTime)
				finishing.XValues = append(finishing.XValues, float64(len(finishing.XValues)+1))
				finishing.YValues = append(finishing.YValues, float64(finishingDuration))
			}
			return errors.EnsureStack(s.marshaller.Marshal(w, ci))
		})
	}, prefix...); err != nil {
		return err
	}
	// Reverse the x values since we collect them from newest to oldest.
	// TODO: It would probably be better to support listing jobs in reverse order.
	reverseContinuousSeries(compacting, validating, finishing)
	return collectGraph(tw, "commits-chart.png", "number of commits", []chart.Series{compacting, validating, finishing}, prefix...)
}

func reverseContinuousSeries(series ...chart.ContinuousSeries) {
	for _, s := range series {
		for i := 0; i < len(s.XValues)/2; i++ {
			s.XValues[i], s.XValues[len(s.XValues)-1-i] = s.XValues[len(s.XValues)-1-i], s.XValues[i]
		}
	}
}

func collectGraph(tw *tar.Writer, name, XAxisName string, series []chart.Series, prefix ...string) error {
	return collectDebugFile(tw, name, "", func(w io.Writer) error {
		graph := chart.Chart{
			Title: name,
			TitleStyle: chart.Style{
				Show: true,
			},
			XAxis: chart.XAxis{
				Name: XAxisName,
				NameStyle: chart.Style{
					Show: true,
				},
				Style: chart.Style{
					Show: true,
				},
				ValueFormatter: func(v interface{}) string {
					return fmt.Sprintf("%.0f", v.(float64))
				},
			},
			YAxis: chart.YAxis{
				Name: "time (seconds)",
				NameStyle: chart.Style{
					Show: true,
				},
				Style: chart.Style{
					Show:                true,
					TextHorizontalAlign: chart.TextHorizontalAlignLeft,
				},
				ValueFormatter: func(v interface{}) string {
					seconds := v.(float64) / float64(time.Second)
					return fmt.Sprintf("%.3f", seconds)
				},
			},
			Series: series,
		}
		graph.Elements = []chart.Renderable{
			chart.Legend(&graph),
		}
		return errors.EnsureStack(graph.Render(chart.PNG, w))
	}, prefix...)
}

func collectGoInfo(tw *tar.Writer, prefix ...string) error {
	return collectDebugFile(tw, "go_info", "txt", func(w io.Writer) error {
		fmt.Fprintf(w, "build info: ")
		info, ok := runtimedebug.ReadBuildInfo()
		if ok {
			fmt.Fprintf(w, "%s", info.String())
		} else {
			fmt.Fprint(w, "<no build info>")
		}
		fmt.Fprintf(w, "GOOS: %v\nGOARCH: %v\nGOMAXPROCS: %v\nNumCPU: %v\n", runtime.GOOS, runtime.GOARCH, runtime.GOMAXPROCS(0), runtime.NumCPU())
		return nil
	}, prefix...)
}

func (s *debugServer) collectPachdVersion(ctx context.Context, tw *tar.Writer, pachClient *client.APIClient, prefix ...string) error {
	return collectDebugFile(tw, "version", "txt", func(w io.Writer) (retErr error) {
		defer log.Span(ctx, "collectPachdVersion")(log.Errorp(&retErr))
		version, err := pachClient.Version()
		if err != nil {
			return err
		}
		_, err = io.Copy(w, strings.NewReader(version+"\n"))
		return errors.EnsureStack(err)
	}, prefix...)
}

func (s *debugServer) collectDescribe(ctx context.Context, tw *tar.Writer, pod string, prefix ...string) error {
	return collectDebugFile(tw, "describe", "txt", func(output io.Writer) (retErr error) {
		// Gate the total time of "describe".
		ctx, c := context.WithTimeout(ctx, 2*time.Minute)
		defer c()
		defer log.Span(ctx, "collectDescribe")(log.Errorp(&retErr))
		r, w := io.Pipe()
		go func() {
			defer log.Span(ctx, "collectDescribe.backgroundDescribe")()
			// Do the "describe" in the background, because the k8s client doesn't take
			// a context here, and we want the ability to abandon the request.  We leak
			// memory if this runs forever but we return, but that's better than a debug
			// dump that doesn't return.
			pd := describe.PodDescriber{
				Interface: s.env.GetKubeClient(),
			}
			output, err := pd.Describe(s.env.Config().Namespace, pod, describe.DescriberSettings{ShowEvents: true})
			if err != nil {
				w.CloseWithError(errors.EnsureStack(err))
				return
			}
			_, err = w.Write([]byte(output))
			w.CloseWithError(errors.EnsureStack(err))
		}()
		go func() {
			// Close the pipe when the context times out; bounding the time on the
			// io.Copy operation below.
			<-ctx.Done()
			w.CloseWithError(errors.EnsureStack(ctx.Err()))
		}()
		if _, err := io.Copy(output, r); err != nil {
			return errors.EnsureStack(err)
		}
		return nil
	}, prefix...)
}

func (s *debugServer) collectLogs(ctx context.Context, tw *tar.Writer, pod, container string, prefix ...string) error {
	if err := collectDebugFile(tw, "logs", "txt", func(w io.Writer) (retErr error) {
		defer log.Span(ctx, "collectLogs", zap.String("pod", pod), zap.String("container", container))(log.Errorp(&retErr))
		stream, err := s.env.GetKubeClient().CoreV1().Pods(s.env.Config().Namespace).GetLogs(pod, &v1.PodLogOptions{Container: container}).Stream(ctx)
		if err != nil {
			return errors.EnsureStack(err)
		}
		defer func() {
			if err := stream.Close(); retErr == nil {
				retErr = err
			}
		}()
		_, err = io.Copy(w, stream)
		return errors.EnsureStack(err)
	}, prefix...); err != nil {
		return err
	}
	return collectDebugFile(tw, "logs-previous", "txt", func(w io.Writer) (retErr error) {
		defer log.Span(ctx, "collectLogs.previous", zap.String("pod", pod), zap.String("container", container))(log.Errorp(&retErr))
		stream, err := s.env.GetKubeClient().CoreV1().Pods(s.env.Config().Namespace).GetLogs(pod, &v1.PodLogOptions{Container: container, Previous: true}).Stream(ctx)
		if err != nil {
			return errors.EnsureStack(err)
		}
		defer func() {
			if err := stream.Close(); retErr == nil {
				retErr = err
			}
		}()
		_, err = io.Copy(w, stream)
		return errors.EnsureStack(err)
	}, prefix...)
}

func (s *debugServer) hasLoki() bool {
	_, err := s.env.GetLokiClient()
	return err == nil
}

func (s *debugServer) collectLogsLoki(ctx context.Context, tw *tar.Writer, pod, container string, prefix ...string) error {
	if !s.hasLoki() {
		return nil
	}
	return collectDebugFile(tw, "logs-loki", "txt", func(w io.Writer) (retErr error) {
		defer log.Span(ctx, "collectLogsLoki", zap.String("pod", pod), zap.String("container", container))(log.Errorp(&retErr))
		queryStr := `{pod="` + pod
		if container != "" {
			queryStr += `", container="` + container
		}
		queryStr += `"}`
		logs, err := s.queryLoki(ctx, queryStr)
		if err != nil {
			return errors.EnsureStack(err)
		}

		var cursor *loki.LabelSet
		for _, entry := range logs {
			// Print the stream labels in %v format whenever they are different from the
			// previous line.  The pointer comparison is a fast path to avoid
			// reflect.DeepEqual when both log lines are from the same chunk of logs
			// returned by Loki.
			if cursor != entry.Labels && !reflect.DeepEqual(cursor, entry.Labels) {
				if _, err := fmt.Fprintf(w, "%v\n", entry.Labels); err != nil {
					return errors.EnsureStack(err)
				}
			}
			cursor = entry.Labels

			// Then the line itself.
			if _, err := fmt.Fprintf(w, "%s\n", entry.Entry.Line); err != nil {
				return errors.EnsureStack(err)
			}
		}
		return nil
	}, prefix...)
}

// collectDump is done on the pachd side, AND by the workers by them.
func collectDump(ctx context.Context, tw *tar.Writer, _ *client.APIClient, prefix ...string) (retErr error) {
	defer log.Span(ctx, "collectDump", zap.Strings("prefix", prefix))(log.Errorp(&retErr))

	// Collect go info.
	if err := collectGoInfo(tw, prefix...); err != nil {
		return err
	}

	// Goroutine profile.
	if err := collectProfile(ctx, tw, &debug.Profile{Name: "goroutine"}, prefix...); err != nil {
		return err
	}

	// Heap profile.
	if err := collectProfile(ctx, tw, &debug.Profile{Name: "heap"}, prefix...); err != nil {
		return err
	}
	return nil
}

func (s *debugServer) collectPipelineDumpFunc(limit int64) collectPipelineFunc {
	return func(ctx context.Context, tw *tar.Writer, pachClient *client.APIClient, pipelineInfo *pps.PipelineInfo, prefix ...string) (retErr error) {
		ctx, end := log.SpanContext(ctx, "collectPipelineDump", zap.Stringer("pipeline", pipelineInfo.GetPipeline()))
		defer end(log.Errorp(&retErr))

		if err := validatePipelineInfo(pipelineInfo); err != nil {
			return errors.Wrap(err, "collectPipelineDumpFunc: invalid pipeline info")
		}
		var errs error
		if err := collectDebugFile(tw, "spec", "json", func(w io.Writer) error {
			fullPipelineInfos, err := pachClient.ListProjectPipelineHistory(pipelineInfo.Pipeline.Project.Name, pipelineInfo.Pipeline.Name, -1, true)
			if err != nil {
				return err
			}
			var pipelineErrs error
			for _, fullPipelineInfo := range fullPipelineInfos {
				if err := s.marshaller.Marshal(w, fullPipelineInfo); err != nil {
					multierr.AppendInto(&pipelineErrs, errors.Wrapf(err, "marshalFullPipelineInfo(%s)", fullPipelineInfo.GetPipeline()))
				}
			}
			return nil
		}, prefix...); err != nil {
			multierr.AppendInto(&errs, errors.Wrap(err, "listProjectPipelineHistory"))
		}
		if err := s.collectCommits(ctx, tw, pachClient, client.NewProjectRepo(pipelineInfo.Pipeline.Project.GetName(), pipelineInfo.Pipeline.Name), limit, prefix...); err != nil {
			multierr.AppendInto(&errs, errors.Wrap(err, "collectCommits"))
		}
		if err := s.collectJobs(tw, pachClient, pipelineInfo.Pipeline, limit, prefix...); err != nil {
			multierr.AppendInto(&errs, errors.Wrap(err, "collectJobs"))
		}
		if s.hasLoki() {
			if err := s.forEachWorkerLoki(ctx, pipelineInfo, func(pod string) (retErr error) {
				ctx, end := log.SpanContext(ctx, "forEachWorkerLoki.worker", zap.String("pod", pod))
				defer end(log.Errorp(&retErr))

				workerPrefix := join(podPrefix, pod)
				if len(prefix) > 0 {
					workerPrefix = join(prefix[0], workerPrefix)
				}
				userPrefix := join(workerPrefix, client.PPSWorkerUserContainerName)
				if err := s.collectLogsLoki(ctx, tw, pod, client.PPSWorkerUserContainerName, userPrefix); err != nil {
					return err
				}
				sidecarPrefix := join(workerPrefix, client.PPSWorkerSidecarContainerName)
				return s.collectLogsLoki(ctx, tw, pod, client.PPSWorkerSidecarContainerName, sidecarPrefix)
			}); err != nil {
				name := "loki"
				if len(prefix) > 0 {
					name = join(prefix[0], name)
				}
				return writeErrorFile(tw, err, name)
			}
		}
		return nil
	}
}

func (s *debugServer) forEachWorkerLoki(ctx context.Context, pipelineInfo *pps.PipelineInfo, cb func(string) error) error {
	pods, err := s.getWorkerPodsLoki(ctx, pipelineInfo)
	if err != nil {
		return err
	}
	var errs error
	for pod := range pods {
		if err := cb(pod); err != nil {
			multierr.AppendInto(&errs, errors.Wrapf(err, "forEachWorkersLoki(%s)", pod))
		}
	}
	return errs
}

// quoteLogQLStreamSelector returns a string quoted as a LogQL stream selector.
// The rules for quoting a LogQL stream selector are documented at
// https://grafana.com/docs/loki/latest/logql/log_queries/#log-stream-selector
// to be the same as for a Prometheus string literal, as documented at
// https://prometheus.io/docs/prometheus/latest/querying/basics/#string-literals.
// This happens to be the same as a Go string, with single or double quotes or
// backticks allowed as enclosing characters.
func quoteLogQLStreamSelector(s string) string {
	return strconv.Quote(s)
}

func (s *debugServer) getWorkerPodsLoki(ctx context.Context, pipelineInfo *pps.PipelineInfo) (map[string]struct{}, error) {
	// This function uses the log querying API and not the label querying API, to bound the the
	// number of workers for a pipeline that we return.  We'll get 30,000 of the most recent
	// logs for each pipeline, and return the names of the workers that contributed to those
	// logs for further inspection.  The alternative would be to get every worker that existed
	// in some time interval, but that results in too much data to inspect.

	queryStr := fmt.Sprintf(`{pipelineProject=%s, pipelineName=%s}`, quoteLogQLStreamSelector(pipelineInfo.Pipeline.Project.Name), quoteLogQLStreamSelector(pipelineInfo.Pipeline.Name))
	pods := make(map[string]struct{})
	logs, err := s.queryLoki(ctx, queryStr)
	if err != nil {
		return nil, err
	}

	for _, l := range logs {
		if l.Labels == nil {
			continue
		}
		labels := *l.Labels
		pod, ok := labels["pod"]
		if !ok {
			log.Debug(ctx, "pod label missing from loki labelset", zap.Any("log", l))
			continue
		}
		pods[pod] = struct{}{}
	}
	return pods, nil
}

const (
	// maxLogs used to be 5,000 but a comment said this was too few logs, so now it's 30,000.
	// If it still seems too small, bump it up again.
	maxLogs = 30000
	// 5,000 is the maximum value of "limit" in queries that Loki seems to accept.
	// We set serverMaxLogs below the actual limit because of a bug in Loki where it hangs when you request too much data from it.
	serverMaxLogs = 1000
)

type lokiLog struct {
	Labels *loki.LabelSet
	Entry  *loki.Entry
}

func (s *debugServer) queryLoki(ctx context.Context, queryStr string) (retResult []lokiLog, retErr error) {
	ctx, finishSpan := log.SpanContext(ctx, "queryLoki", zap.String("queryStr", queryStr))
	defer func() {
		finishSpan(zap.Error(retErr), zap.Int("logs", len(retResult)))
	}()

	sortLogs := func(logs []lokiLog) {
		sort.Slice(logs, func(i, j int) bool {
			return logs[i].Entry.Timestamp.Before(logs[j].Entry.Timestamp)
		})
	}

	c, err := s.env.GetLokiClient()
	if err != nil {
		return nil, errors.EnsureStack(errors.Errorf("get loki client: %v", err))
	}
	// We used to just stream the output, but that results in logs from different streams
	// (stdout/stderr) being randomly interspersed with each other, which is hard to follow when
	// written in text form.  We also used "FORWARD" and basically got the first 30,000 logs
	// this month, instead of the most recent 30,000 logs.  To use "BACKWARD" to get chunks of
	// logs starting with the most recent (whose start time we can't know without asking), we
	// have to either output logs in reverse order, or collect them all and then reverse.  Thus,
	// we just buffer.  30k logs * (200 bytes each + 16 bytes of pointers per line + 24 bytes of
	// time.Time) = 6.8MiB.  That seems totally reasonable to keep in memory, with the benefit
	// of providing lines in chronological order even when the app uses both stdout and stderr.
	var result []lokiLog

	var end time.Time
	start := time.Now().Add(-30 * 24 * time.Hour) // 30 days.  (Loki maximum range is 30 days + 1 hour.)

	for numLogs := 0; (end.IsZero() || start.Before(end)) && numLogs < maxLogs; {
		// Loki requests can hang if the size of the log lines is too big
		ctx, cancel := context.WithTimeout(ctx, time.Minute)
		resp, err := c.QueryRange(ctx, queryStr, serverMaxLogs, start, end, "BACKWARD", 0 /* step */, 0 /* interval */, true /* quiet */)
		cancel()
		if err != nil {
			// Note: the error from QueryRange has a stack.
			if errors.Is(err, context.DeadlineExceeded) {
				log.Debug(ctx, "loki query range timed out", zap.Int("serverMaxLogs", serverMaxLogs), zap.Time("start", start), zap.Time("end", end), zap.Int("logs", len(result)))
				sortLogs(result)
				return result, nil
			}
			return nil, errors.Errorf("query range (query=%v, limit=%v, start=%v, end=%v): %+v", queryStr, serverMaxLogs, start.Format(time.RFC3339), end.Format(time.RFC3339), err)
		}

		streams, ok := resp.Data.Result.(loki.Streams)
		if !ok {
			return nil, errors.Errorf("resp.Data.Result must be of type loki.Streams")
		}

		var readThisIteration int
		for _, s := range streams {
			stream := s // Alias for pointer later.
			for _, e := range stream.Entries {
				entry := e
				numLogs++
				readThisIteration++
				if end.IsZero() || entry.Timestamp.Before(end) {
					// Because end is an "exclusive" range, if we read any logs
					// at all we are guaranteed to not read them in the next
					// iteration.  (If all 5000 logs are at the same time, we
					// still advance past them on the next iteration, which is
					// the best we can do under those circumstances.)
					end = entry.Timestamp
				}
				result = append(result, lokiLog{
					Labels: &stream.Labels,
					Entry:  &entry,
				})
			}
		}
		if readThisIteration == 0 {
			break
		}
		cancel()
	}
	sortLogs(result)
	return result, nil
}

func (s *debugServer) collectJobs(tw *tar.Writer, pachClient *client.APIClient, pipeline *pps.Pipeline, limit int64, prefix ...string) error {
	download := chart.ContinuousSeries{
		Name: "download",
		Style: chart.Style{
			Show:        true,
			StrokeColor: chart.GetDefaultColor(0).WithAlpha(255),
		},
	}
	process := chart.ContinuousSeries{
		Name: "process",
		Style: chart.Style{
			Show:        true,
			StrokeColor: chart.GetDefaultColor(1).WithAlpha(255),
		},
	}
	upload := chart.ContinuousSeries{
		Name: "upload",
		Style: chart.Style{
			Show:        true,
			StrokeColor: chart.GetDefaultColor(2).WithAlpha(255),
		},
	}
	if err := collectDebugFile(tw, "jobs", "json", func(w io.Writer) error {
		// TODO: The limiting should eventually be a feature of list job.
		var count int64
		return pachClient.ListProjectJobF(pipeline.Project.Name, pipeline.Name, nil, -1, false, func(ji *pps.JobInfo) error {
			if count >= limit {
				return errutil.ErrBreak
			}
			count++
			if ji.Stats.DownloadTime != nil && ji.Stats.ProcessTime != nil && ji.Stats.UploadTime != nil {
				downloadDuration, err := types.DurationFromProto(ji.Stats.DownloadTime)
				if err != nil {
					return errors.EnsureStack(err)
				}
				processDuration, err := types.DurationFromProto(ji.Stats.ProcessTime)
				if err != nil {
					return errors.EnsureStack(err)
				}
				uploadDuration, err := types.DurationFromProto(ji.Stats.UploadTime)
				if err != nil {
					return errors.EnsureStack(err)
				}
				download.XValues = append(download.XValues, float64(len(download.XValues)+1))
				download.YValues = append(download.YValues, float64(downloadDuration))
				process.XValues = append(process.XValues, float64(len(process.XValues)+1))
				process.YValues = append(process.YValues, float64(processDuration))
				upload.XValues = append(upload.XValues, float64(len(upload.XValues)+1))
				upload.YValues = append(upload.YValues, float64(uploadDuration))
			}
			return errors.EnsureStack(s.marshaller.Marshal(w, ji))
		})
	}, prefix...); err != nil {
		return err
	}
	// Reverse the x values since we collect them from newest to oldest.
	// TODO: It would probably be better to support listing jobs in reverse order.
	reverseContinuousSeries(download, process, upload)
	return collectGraph(tw, "jobs-chart.png", "number of jobs", []chart.Series{download, process, upload}, prefix...)
}

// collectWorkerDump is something to do for each worker, on the pachd side.
func (s *debugServer) collectWorkerDump(ctx context.Context, tw *tar.Writer, pod *v1.Pod, prefix ...string) (retErr error) {
	defer log.Span(ctx, "collectWorkerDump")(log.Errorp(&retErr))
	// Collect the worker describe output.
	if err := s.collectDescribe(ctx, tw, pod.Name, prefix...); err != nil {
		return err
	}
	// Collect the worker user and storage container logs.
	userPrefix := client.PPSWorkerUserContainerName
	sidecarPrefix := client.PPSWorkerSidecarContainerName
	if len(prefix) > 0 {
		userPrefix = join(prefix[0], userPrefix)
		sidecarPrefix = join(prefix[0], sidecarPrefix)
	}
	var errs error
	if err := s.collectLogs(ctx, tw, pod.Name, client.PPSWorkerUserContainerName, userPrefix); err != nil {
		multierr.AppendInto(&errs, errors.Wrap(err, "userContainerLogs"))
	}
	if err := s.collectLogs(ctx, tw, pod.Name, client.PPSWorkerSidecarContainerName, sidecarPrefix); err != nil {
		multierr.AppendInto(&errs, errors.Wrap(err, "storageContainerLogs"))
	}
	return errs
}

func redirectDump(ctx context.Context, c debug.DebugClient, filter *debug.Filter) (_ io.Reader, retErr error) {
	defer log.Span(ctx, "redirectDump")(log.Errorp(&retErr))
	dumpC, err := c.Dump(ctx, &debug.DumpRequest{Filter: filter})
	if err != nil {
		return nil, errors.EnsureStack(err)
	}
	return grpcutil.NewStreamingBytesReader(dumpC, nil), nil
}

// handleHelmSecret decodes a helm release secret into its various components.  A returned error
// should be written into error.txt.
func handleHelmSecret(ctx context.Context, tw *tar.Writer, secret v1.Secret) error {
	var errs error
	name := secret.Name
	if got, want := string(secret.Type), "helm.sh/release.v1"; got != want {
		return errors.Errorf("helm-owned secret of unknown version; got %v want %v", got, want)
	}
	if secret.Data == nil {
		log.Info(ctx, "skipping helm secret with no data", zap.String("secretName", name))
		return nil
	}
	releaseData, ok := secret.Data["release"]
	if !ok {
		log.Info(ctx, "secret data doesn't have a release key", zap.String("secretName", name), zap.Strings("keys", maps.Keys(secret.Data)))
		return nil
	}

	// There are a few labels that helm adds that are interesting (the "release
	// status").  Grab those and write to metadata.json.
	if err := collectDebugFile(tw, path.Join("helm", name, "metadata"), "json", func(w io.Writer) error {
		text, err := json.Marshal(secret.ObjectMeta)
		if err != nil {
			return errors.Wrap(err, "marshal ObjectMeta")
		}
		if _, err := w.Write(text); err != nil {
			return errors.Wrap(err, "write ObjectMeta json")
		}
		return nil
	}); err != nil {
		// We can still try to get the release JSON if the metadata doesn't marshal
		// or write cleanly.
		multierr.AppendInto(&errs, errors.Wrapf(err, "%v: collect metadata", name))
	}

	// Get the text of the release and write it to release.json.
	var releaseJSON []byte
	if err := collectDebugFile(tw, path.Join("helm", name, "release"), "json", func(w io.Writer) error {
		// The helm release data is base64-encoded gzipped JSON-marshalled protobuf.
		// The base64 encoding is IN ADDITION to the base64 encoding that k8s does
		// when serving secrets through the API; client-go removes that for us.
		var err error
		releaseJSON, err = base64.StdEncoding.DecodeString(string(releaseData))
		if err != nil {
			releaseJSON = nil
			return errors.Wrap(err, "decode base64")
		}

		// Older versions of helm do not compress the data; this check is for the
		// gzip header; if deteced, decompress.
		if len(releaseJSON) > 3 && bytes.Equal(releaseJSON[0:3], []byte{0x1f, 0x8b, 0x08}) {
			gz, err := gzip.NewReader(bytes.NewReader(releaseJSON))
			if err != nil {
				return errors.Wrap(err, "create gzip reader")
			}
			releaseJSON, err = io.ReadAll(gz)
			if err != nil {
				return errors.Wrap(err, "decompress")
			}
			if err := gz.Close(); err != nil {
				return errors.Wrap(err, "close gzip reader")
			}
		}

		// Write out the raw release JSON, so fields we don't pick out in the next
		// phase can be parsed out by the person analyzing the dump if necessary.
		if _, err := w.Write(releaseJSON); err != nil {
			return errors.Wrap(err, "write release")
		}
		return nil
	}); err != nil {
		multierr.AppendInto(&errs, errors.Wrapf(err, "%v: collect release", name))
		// The next steps need the release JSON, so we have to give up if any of
		// this failed.  Technically if the write fails, we could continue, but it
		// doesn't seem worth the effort because the next writes are also likely to
		// fail.
		return errs
	}

	// Unmarshal the release JSON, and start writing files for each interesting part.
	var release struct {
		// Config is the merged values.yaml that aren't in the chart.
		Config map[string]any `json:"config"`
		// Manifest is the rendered manifest that was applied to the cluster.
		Manifest string `json:"manifest"`
	}
	if err := json.Unmarshal(releaseJSON, &release); err != nil {
		multierr.AppendInto(&errs, errors.Wrapf(err, "%v: unmarshal release json", name))
		return errs
	}

	// Write the manifest YAML.
	if err := collectDebugFile(tw, path.Join("helm", name, "manifest"), "yaml", func(w io.Writer) error {
		// Helm adds a newline at the end of the manifest.
		if _, err := fmt.Fprintf(w, "%s", release.Manifest); err != nil {
			return errors.Wrap(err, "print manifest")
		}
		return nil
	}); err != nil {
		multierr.AppendInto(&errs, errors.Wrapf(err, "%v: write manifest.yaml", name))
		// We can try the next step if this fails.
	}

	// Write values.yaml.
	if err := collectDebugFile(tw, path.Join("helm", name, "values"), "yaml", func(w io.Writer) error {
		b, err := yaml.Marshal(release.Config)
		if err != nil {
			return errors.Wrap(err, "marshal values to yaml")
		}
		if _, err := w.Write(b); err != nil {
			return errors.Wrap(err, "print values")
		}
		return nil
	}); err != nil {
		multierr.AppendInto(&errs, errors.Wrapf(err, "%v: write values.yaml", name))
	}
	return errs
}

func (s *debugServer) helmReleases(ctx context.Context, tw *tar.Writer) (retErr error) {
	ctx, end := log.SpanContext(ctx, "collectHelmReleases")
	defer end(log.Errorp(&retErr))
	// Helm stores release data in secrets by default.  Users can override this by exporting
	// HELM_DRIVER and using something else, in which case, we won't get anything here.
	// https://helm.sh/docs/topics/advanced/#storage-backends
	secrets, err := s.env.GetKubeClient().CoreV1().Secrets(s.env.Config().Namespace).List(ctx, metav1.ListOptions{
		LabelSelector: "owner=helm",
	})
	if err != nil {
		return errors.EnsureStack(err)
	}
	var writeErrs error
	for _, secret := range secrets.Items {
		if err := handleHelmSecret(ctx, tw, secret); err != nil {
			if err := writeErrorFile(tw, err, path.Join("helm", secret.Name)); err != nil {
				multierr.AppendInto(&writeErrs, errors.Wrapf(err, "%v: write error.txt", secret.Name))
				continue
			}
		}
	}
	return writeErrs
}

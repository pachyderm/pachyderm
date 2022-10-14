package server

import (
	"archive/tar"
	"context"
	"fmt"
	"io"
	"math"
	"os"
	"reflect"
	"runtime"
	runtimedebug "runtime/debug"
	"runtime/pprof"
	"sort"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/wcharczuk/go-chart"
	v1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	describe "k8s.io/kubectl/pkg/describe"

	"github.com/gogo/protobuf/jsonpb"
	"github.com/gogo/protobuf/types"

	"github.com/pachyderm/pachyderm/v2/src/client"
	"github.com/pachyderm/pachyderm/v2/src/debug"
	"github.com/pachyderm/pachyderm/v2/src/internal/clientsdk"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/errutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/grpcutil"
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
}

// NewDebugServer creates a new server that serves the debug api over GRPC
func NewDebugServer(env serviceenv.ServiceEnv, name string, sidecarClient *client.APIClient, db *pachsql.DB) debug.DebugServer {
	return &debugServer{
		env:           env,
		name:          name,
		sidecarClient: sidecarClient,
		marshaller:    &jsonpb.Marshaler{Indent: "  "},
		database:      db,
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
					if err := collectPachd(ctx, tw, pachClient, pachdContainerPrefix); err != nil {
						return err
					}
					if wantDatabase {
						return s.collectDatabaseDump(ctx, tw, databasePrefix)
					}
				case *debug.Filter_Pipeline:
					pipelineInfo, err := pachClient.InspectPipeline(f.Pipeline.Name, true)
					if err != nil {
						return err
					}
					return s.handlePipelineRedirect(ctx, tw, pachClient, pipelineInfo, collectPipeline, collectWorker, redirect)
				case *debug.Filter_Worker:
					if f.Worker.Redirected {
						// Collect the storage container.
						if s.sidecarClient == nil {
							return collect(ctx, tw, pachClient, client.PPSWorkerSidecarContainerName)
						}
						// Collect the user container.
						if err := collect(ctx, tw, pachClient, client.PPSWorkerUserContainerName); err != nil {
							return err
						}
						// Redirect to the storage container.
						r, err := redirect(ctx, s.sidecarClient.DebugClient, filter)
						if err != nil {
							return err
						}
						return collectDebugStream(tw, r)
					}
					pod, err := s.env.GetKubeClient().CoreV1().Pods(s.env.Config().Namespace).Get(ctx, f.Worker.Pod, metav1.GetOptions{})
					if err != nil {
						return errors.EnsureStack(err)
					}
					return s.handleWorkerRedirect(ctx, tw, pod, collectWorker, redirect)
				case *debug.Filter_Database:
					if wantDatabase {
						return s.collectDatabaseDump(ctx, tw, databasePrefix)
					}
				}
			}
			// No filter, return everything
			if err := collectPachd(ctx, tw, pachClient, pachdContainerPrefix); err != nil {
				return err
			}
			pipelineInfos, err := pachClient.ListPipeline(true)
			if err != nil {
				return err
			}
			for _, pipelineInfo := range pipelineInfos {
				if err := s.handlePipelineRedirect(ctx, tw, pachClient, pipelineInfo, collectPipeline, collectWorker, redirect); err != nil {
					return err
				}
			}
			if wantAppLogs {
				// All other pachyderm apps (console, pg-bouncer, etcd, etc.).
				if err := s.appLogs(ctx, tw); err != nil {
					return err
				}
			}
			if wantDatabase {
				if err := s.collectDatabaseDump(ctx, tw, databasePrefix); err != nil {
					return err
				}
			}
			return nil
		})
	})
}

func (s *debugServer) appLogs(ctx context.Context, tw *tar.Writer) error {
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
	for _, pod := range pods.Items {
		podPrefix := join(pod.Labels["app"], pod.Name)
		if err := s.collectDescribe(tw, pod.Name, podPrefix); err != nil {
			return err
		}
		for _, container := range pod.Spec.Containers {
			prefix := join(podPrefix, container.Name)
			if err := s.collectLogs(ctx, tw, pod.Name, container.Name, prefix); err != nil {
				return err
			}
			if err := s.collectLogsLoki(ctx, tw, pod.Name, container.Name, prefix); err != nil {
				return err
			}
		}
	}
	return nil
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
	prefix := join(pipelinePrefix, pipelineInfo.Pipeline.Name)
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
		return errors.Errorf("no worker pods found for pipeline %v", pipelineInfo.Pipeline.Name)
	}
	for _, pod := range pods {
		if err := cb(&pod); err != nil {
			return err
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
						"app": ppsutil.PipelineRcName(pipelineInfo.Pipeline.Name, pipelineInfo.Version),
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
			log.Errorf("errored closing worker client: %v", err)
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
		collectProfileFunc(request.Profile),
		nil,
		nil,
		redirectProfileFunc(request.Profile),
		collectProfileFunc(request.Profile),
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

func writeProfile(ctx context.Context, w io.Writer, profile *debug.Profile) error {
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
		collectBinary,
		nil,
		nil,
		redirectBinary,
		collectBinary,
		false,
		false,
	)
}

func collectBinary(_ context.Context, tw *tar.Writer, _ *client.APIClient, prefix ...string) error {
	return collectDebugFile(tw, "binary", "", func(w io.Writer) (retErr error) {
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
	pachClient := s.env.GetPachClient(server.Context())
	return s.handleRedirect(
		pachClient,
		server,
		request.Filter,
		s.collectPachdDumpFunc(request.Limit),
		s.collectPipelineDumpFunc(request.Limit),
		s.collectWorkerDump,
		redirectDump,
		collectDump,
		true,
		true,
	)
}

func (s *debugServer) collectPachdDumpFunc(limit int64) collectFunc {
	return func(ctx context.Context, tw *tar.Writer, pachClient *client.APIClient, prefix ...string) error {
		// Collect input repos.
		if err := s.collectInputRepos(ctx, tw, pachClient, limit); err != nil {
			return err
		}
		// Collect the pachd version.
		if err := s.collectPachdVersion(tw, pachClient, prefix...); err != nil {
			return err
		}
		// Collect go info.
		if err := s.collectGoInfo(tw, prefix...); err != nil {
			return err
		}
		// Collect the pachd describe output.
		if err := s.collectDescribe(tw, s.name, prefix...); err != nil {
			return err
		}
		// Collect the pachd container logs.
		if err := s.collectLogs(ctx, tw, s.name, "pachd", prefix...); err != nil {
			return err
		}
		// Collect the pachd container logs from loki.
		if err := s.collectLogsLoki(ctx, tw, s.name, "pachd", prefix...); err != nil {
			return err
		}
		// Collect the pachd container dump.
		return collectDump(ctx, tw, pachClient, prefix...)
	}
}

func (s *debugServer) collectInputRepos(ctx context.Context, tw *tar.Writer, pachClient *client.APIClient, limit int64) error {
	repoInfos, err := pachClient.ListRepo()
	if err != nil {
		return err
	}
	for _, repoInfo := range repoInfos {
		if _, err := pachClient.InspectPipeline(repoInfo.Repo.Name, true); err != nil {
			if errutil.IsNotFoundError(err) {
				repoPrefix := join("source-repos", repoInfo.Repo.Name)
				return s.collectCommits(ctx, tw, pachClient, repoInfo.Repo, limit, repoPrefix)
			}
			return err
		}
	}
	return nil
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
		return clientsdk.ForEachCommit(client, func(ci *pfs.CommitInfo) error {
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

func (s *debugServer) collectGoInfo(tw *tar.Writer, prefix ...string) error {
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

func (s *debugServer) collectPachdVersion(tw *tar.Writer, pachClient *client.APIClient, prefix ...string) error {
	return collectDebugFile(tw, "version", "txt", func(w io.Writer) error {
		version, err := pachClient.Version()
		if err != nil {
			return err
		}
		_, err = io.Copy(w, strings.NewReader(version+"\n"))
		return errors.EnsureStack(err)
	}, prefix...)
}

func (s *debugServer) collectDescribe(tw *tar.Writer, pod string, prefix ...string) error {
	return collectDebugFile(tw, "describe", "txt", func(w io.Writer) error {
		pd := describe.PodDescriber{
			Interface: s.env.GetKubeClient(),
		}
		output, err := pd.Describe(s.env.Config().Namespace, pod, describe.DescriberSettings{ShowEvents: true})
		if err != nil {
			return errors.EnsureStack(err)
		}
		_, err = w.Write([]byte(output))
		return errors.EnsureStack(err)
	}, prefix...)
}

func (s *debugServer) collectLogs(ctx context.Context, tw *tar.Writer, pod, container string, prefix ...string) error {
	if err := collectDebugFile(tw, "logs", "txt", func(w io.Writer) (retErr error) {
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

func (s *debugServer) collectLogsLoki(ctx context.Context, tw *tar.Writer, pod, container string, prefix ...string) error {
	if s.env.Config().LokiHost == "" {
		return nil
	}
	return collectDebugFile(tw, "logs-loki", "txt", func(w io.Writer) error {
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

func collectDump(ctx context.Context, tw *tar.Writer, _ *client.APIClient, prefix ...string) error {
	if err := collectProfile(ctx, tw, &debug.Profile{Name: "goroutine"}, prefix...); err != nil {
		return err
	}
	return collectProfile(ctx, tw, &debug.Profile{Name: "heap"}, prefix...)
}

func (s *debugServer) collectPipelineDumpFunc(limit int64) collectPipelineFunc {
	return func(ctx context.Context, tw *tar.Writer, pachClient *client.APIClient, pipelineInfo *pps.PipelineInfo, prefix ...string) error {
		if err := collectDebugFile(tw, "spec", "json", func(w io.Writer) error {
			fullPipelineInfos, err := pachClient.ListPipelineHistory(pipelineInfo.Pipeline.Name, -1, true)
			if err != nil {
				return err
			}
			for _, fullPipelineInfo := range fullPipelineInfos {
				if err := s.marshaller.Marshal(w, fullPipelineInfo); err != nil {
					return errors.EnsureStack(err)
				}
			}
			return nil
		}, prefix...); err != nil {
			return err
		}
		if err := s.collectCommits(ctx, tw, pachClient, client.NewRepo(pipelineInfo.Pipeline.Name), limit, prefix...); err != nil {
			return err
		}
		if err := s.collectJobs(tw, pachClient, pipelineInfo.Pipeline.Name, limit, prefix...); err != nil {
			return err
		}
		if s.env.Config().LokiHost != "" {
			if err := s.forEachWorkerLoki(ctx, pipelineInfo, func(pod string) error {
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
	for pod := range pods {
		if err := cb(pod); err != nil {
			return err
		}
	}
	return nil
}

func (s *debugServer) getWorkerPodsLoki(ctx context.Context, pipelineInfo *pps.PipelineInfo) (map[string]struct{}, error) {
	// This function uses the log querying API and not the label querying API, to bound the the
	// number of workers for a pipeline that we return.  We'll get 30,000 of the most recent
	// logs for each pipeline, and return the names of the workers that contributed to those
	// logs for further inspection.  The alternative would be to get every worker that existed
	// in some time interval, but that results in too much data to inspect.

	queryStr := `{pipelineName="` + pipelineInfo.Pipeline.Name + `"}`
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
			return nil, errors.Errorf("pod label missing from loki label set")
		}
		pods[pod] = struct{}{}
	}
	return pods, nil
}

// This used to be 5,000 but a comment said this was too few logs, so now it's 30,000.  If it still
// seems too small, bump it up again.  5,000 remains the maximum value of "limit" in queries that
// Loki seems to accept.
const (
	maxLogs       = 30000
	serverMaxLogs = 5000
)

type lokiLog struct {
	Labels *loki.LabelSet
	Entry  *loki.Entry
}

func (s *debugServer) queryLoki(ctx context.Context, queryStr string) ([]lokiLog, error) {
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
		resp, err := c.QueryRange(ctx, queryStr, serverMaxLogs, start, end, "BACKWARD", 0, 0, true)
		if err != nil {
			// Note: the error from QueryRange has a stack.
			return nil, errors.Errorf("query range (query=%v, maxLogs=%v, start=%v, end=%v): %+v", queryStr, maxLogs, start.Format(time.RFC3339), end.Format(time.RFC3339), err)
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
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].Entry.Timestamp.Before(result[j].Entry.Timestamp)
	})
	return result, nil
}

func (s *debugServer) collectJobs(tw *tar.Writer, pachClient *client.APIClient, pipelineName string, limit int64, prefix ...string) error {
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
		return pachClient.ListJobF(pipelineName, nil, -1, false, func(ji *pps.JobInfo) error {
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

func (s *debugServer) collectWorkerDump(ctx context.Context, tw *tar.Writer, pod *v1.Pod, prefix ...string) error {
	// Collect the worker describe output.
	if err := s.collectDescribe(tw, pod.Name, prefix...); err != nil {
		return err
	}
	// Collect go info.
	if err := s.collectGoInfo(tw, prefix...); err != nil {
		return err
	}
	// Collect the worker user and storage container logs.
	userPrefix := client.PPSWorkerUserContainerName
	sidecarPrefix := client.PPSWorkerSidecarContainerName
	if len(prefix) > 0 {
		userPrefix = join(prefix[0], userPrefix)
		sidecarPrefix = join(prefix[0], sidecarPrefix)
	}
	if err := s.collectLogs(ctx, tw, pod.Name, client.PPSWorkerUserContainerName, userPrefix); err != nil {
		return err
	}
	return s.collectLogs(ctx, tw, pod.Name, client.PPSWorkerSidecarContainerName, sidecarPrefix)
}

func redirectDump(ctx context.Context, c debug.DebugClient, filter *debug.Filter) (io.Reader, error) {
	dumpC, err := c.Dump(ctx, &debug.DumpRequest{Filter: filter})
	if err != nil {
		return nil, errors.EnsureStack(err)
	}
	return grpcutil.NewStreamingBytesReader(dumpC, nil), nil
}
